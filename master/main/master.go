package main

import (
	"distributed_contrab/master"
	"flag"
	"fmt"
	"runtime"
	"time"
)

var (
	confFile string
)

//解析命令行参数
func initArgs() {
	//master -config ./master.json
	//master -h
	flag.StringVar(&confFile, "config", "./master/main/master.json", "specify master.json")
	flag.Parse()
}

func initEnv() {
	//设置线程数为cpu逻辑核心数
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)

	//初始化命令行参数
	initArgs()

	//初始化线程
	initEnv()

	//初始化配置,json配置文件放置到GOPATH目录下
	err = master.InitConfig(confFile)
	if err != nil {
		goto ERR
	}

	// 初始化服务发现模块
	if err = master.InitWorkerMgr(); err != nil {
		goto ERR
	}

	//初始化任务管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	//启动api http服务,用于接收web页面请求
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(time.Second)
	}
ERR:
	fmt.Println(err)
}
