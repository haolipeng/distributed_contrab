package main

import (
	"fmt"
	"github.com/gorhill/cronexpr"
	"time"
)

func main() {
	//每5分钟执行一次
	cronString := "*/5 * * * * * *"
	expr, err := cronexpr.Parse(cronString)
	if err != nil {
		fmt.Println(err)
		return
	}

	// 当前时间
	curTime := time.Now()

	// 下次调度的时间
	nextTime := expr.Next(curTime)

	//等待定时器超时
	time.AfterFunc(nextTime.Sub(curTime), func() {
		fmt.Println("被调度了")
	})

	time.Sleep(10 * time.Second)
}
