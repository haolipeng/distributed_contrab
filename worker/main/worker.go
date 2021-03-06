package main

import (
	"distributed_contrab/worker"
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
	//worker -config ./worker.json
	//worker -h
	flag.StringVar(&confFile, "config", "./worker/main/worker.json", "specify worker.json")
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
	err = worker.InitConfig(confFile)
	if err != nil {
		goto ERR
	}

	// 初始化日志管理器
	err = worker.InitLogSink()
	if err != nil {
		goto ERR
	}

	//初始化注册节点
	err = worker.InitRegister()
	if err != nil {
		goto ERR
	}

	//初始化执行器
	err = worker.InitExecutor()
	if err != nil {
		goto ERR
	}

	//初始化调度器
	err = worker.InitScheduler()
	if err != nil {
		goto ERR
	}

	//初始化任务管理
	err = worker.InitJobMgr()
	if err != nil {
		goto ERR
	}

	for {
		time.Sleep(time.Second)
	}
ERR:
	fmt.Println(err)
}
