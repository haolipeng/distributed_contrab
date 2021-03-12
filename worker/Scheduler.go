package worker

import (
	"distributed_contrab/common"
	"fmt"
	"time"
)

//任务调度
type Scheduler struct {
	jobEventChan      chan *common.JobEvent
	jobPlanTable      map[string]*common.JobSchedulerPlan //任务调度计划表
	jobExecutingTable map[string]*common.JobExecuteInfo   //任务执行状态表
	jobResultChan     chan *common.JobExecuteResult       //任务结果队列 类型:channel
}

var (
	G_scheduler *Scheduler
)

//初始化调度器
//error返回调度器是否初始化成功
func InitScheduler() error {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent),
		jobPlanTable:      make(map[string]*common.JobSchedulerPlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 100), //默认容量为100
	}

	//启动调度协程
	go G_scheduler.schedulerLoop()

	return nil
}

func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

//将任务执行结果添加到队列中，等待集中处理
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}

//尝试执行任务，此时可能任务处于执行状态
func (scheduler *Scheduler) tryStartJob(jobPlan *common.JobSchedulerPlan) {
	//判断任务是否处于执行状态
	if _, isExist := scheduler.jobExecutingTable[jobPlan.Job.Name]; isExist {
		fmt.Println("任务执行尚未退出，跳出此次执行")
		return
	}

	//不存在，则创建，将JobSchedulerPlan转换JobExecuteInfo
	executeJobInfo := common.BuildJobExecuteInfo(jobPlan)

	//保存执行状态
	scheduler.jobExecutingTable[jobPlan.Job.Name] = executeJobInfo

	// 开始执行任务
	fmt.Println("执行任务:", executeJobInfo.Job.Name)
	G_Executor.ExecutorJob(executeJobInfo)
}

//处理任务执行的结果
func (scheduler *Scheduler) handleJobExecuteResult(result *common.JobExecuteResult) {
	//删除执行状态
	if result.JobInfo == nil {
		fmt.Println("handleJobExecuteResult function result.JobInfo is nil pointer")
	}
	delete(scheduler.jobExecutingTable, result.JobInfo.Job.Name)
	fmt.Printf("任务 %s 执行完成，从jobExecutingTable中删除运行状态\n", result.JobInfo.Job.Name)
	//TODO:生成执行日志
}

//处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE: //新增和修改任务
		var jobSchedulePlan *common.JobSchedulerPlan
		var err error
		if jobSchedulePlan, err = common.BuildJobSchedulerPlan(jobEvent.JobInfo); err != nil {
			return
		}
		//更新任务调度计划表
		scheduler.jobPlanTable[jobEvent.JobInfo.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE: //删除任务
		//判断任务是否存在
		if _, ok := scheduler.jobPlanTable[jobEvent.JobInfo.Name]; ok {
			delete(scheduler.jobPlanTable, jobEvent.JobInfo.Name)
		}
	case common.JOB_EVENT_KILL: //强杀任务
		// 判断任务是否在运行状态，处于运行状态则取消命令的执行
		if jobExecuteInfo, ok := scheduler.jobExecutingTable[jobEvent.JobInfo.Name]; ok {
			jobExecuteInfo.CancelFunc()
		}
	}
}

// 重新计算任务调度状态
func (scheduler *Scheduler) TryScheduler() time.Duration {
	// 获取当前时间
	curTime := time.Now()
	var nearTime *time.Time

	// 如果任务计划表为空，则随便睡眠多久，默认先睡眠
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter := 1 * time.Second
		return scheduleAfter
	}

	// 遍历任务计划调度表中所有任务
	for _, jobPlan := range scheduler.jobPlanTable {
		// 任务时间小于等于当前时间，则触发任务执行
		if jobPlan.NextTime.Before(curTime) || jobPlan.NextTime.Equal(curTime) {
			// 尝试开始执行任务
			scheduler.tryStartJob(jobPlan)

			// 重新计算过期时间
			//fmt.Printf("job name:%s execute time:%s\n", jobPlan.Job, jobPlan.NextTime.String())
			jobPlan.NextTime = jobPlan.Expr.Next(curTime)
		}

		// 统计最近一个要过期的任务的时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 下次调度间隔(最近要执行的任务的调度时间 - 当前时间)
	scheduleAfter := nearTime.Sub(curTime)
	return scheduleAfter
}

//调度协程
func (scheduler *Scheduler) schedulerLoop() {
	// 初始化一次(1秒)
	var schedulerAfter time.Duration
	schedulerAfter = scheduler.TryScheduler()

	schedulerTimer := time.NewTimer(schedulerAfter)
	for {
		var jobEvent *common.JobEvent
		var jobExecuteResult *common.JobExecuteResult
		select {
		//监控任务变化事件
		case jobEvent = <-scheduler.jobEventChan:
			//对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C:
		// 处理任务执行结果
		case jobExecuteResult = <-scheduler.jobResultChan:
			scheduler.handleJobExecuteResult(jobExecuteResult)
		}

		//调度一次任务
		schedulerAfter = scheduler.TryScheduler()
		schedulerTimer.Reset(schedulerAfter)
	}
}
