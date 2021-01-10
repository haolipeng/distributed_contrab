package master

import (
	"distributed_contrab/common"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	//单例对象
	G_apiServer *apiServer
)

//任务的http接口
type apiServer struct {
	httpServer *http.Server
}

//保存任务接口
//POST job={"name":"job1", "command":"echo hello", "cronExpr":"*****"}
func handleJobSave(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		jobContent  string
		job         common.Job
		oldJob      *common.Job
		respContent []byte
	)

	//1.解析post表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	//2.取表单中的job字段
	jobContent = r.PostForm.Get("job")

	fmt.Printf("/job/save : %s\n", jobContent)

	//3.反序列化job
	if err = json.Unmarshal([]byte(jobContent), &job); err != nil {
		goto ERR
	}

	//4.保存到etcd中
	if oldJob, err = G_JobMgr.JobSave(&job); err != nil {
		goto ERR
	}

	//5.返回正常应答
	if respContent, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(respContent)
	}

	return

ERR:
	//6.返回异常应答
	if respContent, err = common.BuildResponse(-1, "failed", nil); err == nil {
		w.Write(respContent)
	}
}

//枚举所有任务(需要补充测试用例)
func handleJobList(w http.ResponseWriter, r *http.Request) {
	var (
		jobList     []*common.Job
		err         error
		respContent []byte
	)

	if jobList, err = G_JobMgr.JobList(); err != nil {
		goto ERR
	}

	//5.返回正常应答
	if respContent, err = common.BuildResponse(0, "success", jobList); err == nil {
		w.Write(respContent)
	}

	return

ERR:
	//6.返回异常应答
	if respContent, err = common.BuildResponse(-1, "failed", nil); err == nil {
		w.Write(respContent)
	}
}

//删除任务接口  /job/delete  name=job1
func handleJobDelete(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		name        string
		oldJob      *common.Job
		respContent []byte
	)

	//1.解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	//2.获取待删除的任务名
	name = r.PostForm.Get("name")

	//3.从etcd中删除任务
	if oldJob, err = G_JobMgr.JobDelete(name); err != nil {
		goto ERR
	}

	//4.返回正常应答
	if respContent, err = common.BuildResponse(0, "success", oldJob); err == nil {
		w.Write(respContent)
	}

	return
ERR:
	//5.返回异常应答
	if respContent, err = common.BuildResponse(-1, "failed", nil); err == nil {
		w.Write(respContent)
	}
}

//初始化服务
func InitApiServer() error {
	var (
		mux        *http.ServeMux
		listen     net.Listener
		httpServer *http.Server
		err        error
	)

	//设置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)

	//启动TCP侦听
	if listen, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		return err
	}

	//创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	//赋值单例
	G_apiServer = &apiServer{
		httpServer: httpServer,
	}

	//开始提供服务
	go httpServer.Serve(listen)

	return err
}
