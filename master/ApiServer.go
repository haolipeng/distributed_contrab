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

	fmt.Printf("任务保存接口/job/save: %s\n", jobContent)

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

	fmt.Println("获取任务列表接口/job/list")

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

	fmt.Printf("删除任务接口/job/delete: %s\n", name)

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

// 查询任务日志
func handleJobLog(resp http.ResponseWriter, req *http.Request) {
	var (
		err        error
		name       string // 任务名字
		skipParam  string // 从第几条开始
		limitParam string // 返回多少条
		skip       int
		limit      int
		logArr     []*common.JobLog
		bytes      []byte
	)

	// 解析GET参数
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 获取请求参数 /job/log?name=job10&skip=0&limit=10
	name = req.Form.Get("name")
	skipParam = req.Form.Get("skip")
	limitParam = req.Form.Get("limit")
	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logArr, err = G_logMgr.ListLog(name, skip, limit); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

//强杀任务  /job/kill  name=job1
func handleJobKill(w http.ResponseWriter, r *http.Request) {
	var (
		err         error
		name        string
		respContent []byte
	)

	//1.解析POST表单
	if err = r.ParseForm(); err != nil {
		goto ERR
	}

	//2.获取待删除的任务名
	name = r.PostForm.Get("name")

	fmt.Printf("强杀任务接口/job/kill: %s\n", name)

	//3.从etcd中删除任务
	if err = G_JobMgr.JobKill(name); err != nil {
		goto ERR
	}

	//4.返回正常应答
	if respContent, err = common.BuildResponse(0, "success", ""); err == nil {
		w.Write(respContent)
	}

	return
ERR:
	//5.返回异常应答
	if respContent, err = common.BuildResponse(-1, "failed", nil); err == nil {
		w.Write(respContent)
	}
}

// 获取健康worker节点列表
func handleWorkerList(resp http.ResponseWriter, req *http.Request) {
	var (
		workerArr []string
		err       error
		bytes     []byte
	)

	if workerArr, err = G_workerMgr.ListWorkers(); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

//初始化gin服务
func InitGinServer() error {
	return nil
}

//初始化服务
func InitApiServer() error {
	var (
		mux        *http.ServeMux
		listen     net.Listener
		httpServer *http.Server
		err        error
		staticdir  http.Dir
	)

	//设置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkerList)

	//静态文件目录
	staticdir = http.Dir(G_config.WebRoot)
	staticHandler := http.FileServer(staticdir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))

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
