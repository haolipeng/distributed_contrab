package master

import (
	"net"
	"net/http"
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

//保存任务接口(暂时未实现)
func handleJobSave(w http.ResponseWriter, r *http.Request) {

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
	if listen, err = net.Listen("tcp", ":8070"); err != nil {
		return err
	}

	//创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
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
