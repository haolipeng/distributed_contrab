package worker

import (
	"distributed_contrab/common"
	"os/exec"
	"time"
)

type Executor struct {
	//TODO:
}

var (
	G_Executor *Executor
)

//在使用单例G_Executor之前, 必须调用此函数
func InitExecutor() error {
	G_Executor = &Executor{}
	return nil
}

//执行任务
func (executor *Executor) ExecutorJob(jobInfo *common.JobExecuteInfo) {
	//执行任务结果结构体
	result := &common.JobExecuteResult{
		Output:      make([]byte, 0), //TODO:这块不写的话，会出问题吗
		ExecuteInfo: jobInfo,
	}

	//TODO: 这块需要分布式锁，成功获取到分布式锁，才能开始执行任务
	// 初始化分布式锁

	result.StartTime = time.Now()

	//真正执行命令，修复windows上使用bash的兼容性bug
	cmd := exec.CommandContext(jobInfo.CancelCtx, "F:\\cygwin64\\bin\\bash.exe", "-c", jobInfo.Job.Command)
	output, err := cmd.CombinedOutput()

	result.Output = output
	result.Err = err
	result.EndTime = time.Now()

	G_scheduler.PushJobResult(result)
}
