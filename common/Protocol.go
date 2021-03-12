package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//任务信息
type Job struct {
	Name     string `json:"name"`     //任务名
	Command  string `json:"command"`  //shell命令
	CronExpr string `json:"cronExpr"` //crontab表达式
}

//任务调度计划结构体
type JobSchedulerPlan struct {
	Job      *Job                 //属于哪个job
	Expr     *cronexpr.Expression //解析后的crontab表达式
	NextTime time.Time            //下次调度时间
}

// 任务执行状态
type JobExecuteInfo struct {
	Job        *Job
	PlanTime   time.Time          //理论上的调度时间
	RealTime   time.Time          //实际的调度时间
	CancelCtx  context.Context    //任务command的上下文
	CancelFunc context.CancelFunc //用户取消command执行的cancel函数
}

// 任务执行结果
type JobExecuteResult struct {
	JobInfo   *JobExecuteInfo
	Output    []byte
	Err       error
	StartTime time.Time
	EndTime   time.Time
}

type JobEvent struct {
	EventType int
	JobInfo   *Job
}

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func BuildJobExecuteInfo(jobPlan *JobSchedulerPlan) *JobExecuteInfo {
	ctx, cancelFunc := context.WithCancel(context.TODO())

	executeInfo := &JobExecuteInfo{
		Job:        jobPlan.Job,
		PlanTime:   jobPlan.NextTime,
		RealTime:   time.Now(),
		CancelCtx:  ctx,
		CancelFunc: cancelFunc,
	}

	return executeInfo
}

func BuildJobSchedulerPlan(job *Job) (*JobSchedulerPlan, error) {
	//解析job的cron表达式
	var (
		expr *cronexpr.Expression
		err  error
	)

	expr, err = cronexpr.Parse(job.CronExpr)
	if err != nil {
		return nil, err
	}

	return &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}, nil
}

func BuildResponse(errno int, msg string, data interface{}) ([]byte, error) {
	var (
		resp    Response
		err     error
		content []byte
	)

	//变量赋值
	resp.Errno = errno
	resp.Msg = msg
	resp.Data = data

	//json序列化
	content, err = json.Marshal(resp)

	return content, err
}

//反序列化job
func UnpackJob(value []byte) (*Job, error) {
	//将字节流反序列化成Job对象
	var job *Job
	var err error
	if err = json.Unmarshal(value, &job); err != nil {
		job = nil
	}

	return job, err
}

func BuildJobEvent(event int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: event,
		JobInfo:   job,
	}
}

func ExtractKillerName(jobKillerKey string) string {
	return strings.TrimPrefix(jobKillerKey, JOB_KILLER_DIR)
}

func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}
