package main

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

//结果结构体
type Result struct {
	err    error
	output []byte
}

func main() {
	//1.创建Result channel通道
	resultCh := make(chan *Result, 2)

	//2.创建带cancel的context上下文
	ctx, cancelFunc := context.WithCancel(context.TODO())

	//2.开协程执行命令，将结果存储到channel通道中
	go func() {
		cmd := exec.CommandContext(ctx, "F:\\cygwin64\\bin\\bash.exe", "-c", "sleep 5;echo hello")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("command exec failed:%s\n", err.Error())
			return
		}

		result := &Result{
			err:    err,
			output: output,
		}

		resultCh <- result
	}()

	//3.取消执行上述命令
	// 继续往下走
	time.Sleep(1 * time.Second)
	cancelFunc()

	//4.从channel中获取执行结果
	newResult := <-resultCh
	fmt.Printf("err:%s, output:%s\n", newResult.err, string(newResult.output))
}
