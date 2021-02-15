package worker

import (
	"distributed_contrab/common"
	"time"
)

//任务调度
type Scheduler struct {
	jobEventChan chan *common.JobEvent
	jobPlanTable map[string]*common.JobSchedulerPlan //任务调度计划表
	//TODO:任务执行表
	//TODO:任务结果队列
}

var (
	G_scheduler *Scheduler
)

func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE: //新增和修改
		var jobSchedulePlan *common.JobSchedulerPlan
		var err error
		if jobSchedulePlan, err = common.BuildJobSchedulerPlan(jobEvent.JobInfo); err != nil {
			return
		}
		//更新任务调度计划表
		scheduler.jobPlanTable[jobEvent.JobInfo.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE: //删除
		//判断任务是否存在
		if _, ok := scheduler.jobPlanTable[jobEvent.JobInfo.Name]; ok {
			delete(scheduler.jobPlanTable, jobEvent.JobInfo.Name)
		}
		//case common.JOB_EVENT_KILL:
		//TODO:强杀任务
	}
}

// 重新计算任务调度状态
func (scheduler *Scheduler) TryScheduler() time.Duration {
	// 获取当前时间
	curTime := time.Now()
	var nearTime *time.Time

	// 如果任务计划表为空，则随便睡眠多久
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter := 1 * time.Second
		return scheduleAfter
	}

	// 遍历所有任务
	for _, jobPlan := range scheduler.jobPlanTable {
		// 任务时间小于等于当前时间
		if jobPlan.NextTime.Before(curTime) || jobPlan.NextTime.Equal(curTime) {
			// TODO:尝试开始执行任务
			// 重新计算过期时间
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
	// 初始化一次(1秒	)
	var schedulerAfter time.Duration
	schedulerAfter = scheduler.TryScheduler()

	schedulerTimer := time.NewTimer(schedulerAfter)
	//1.遍历jobEventChan工作事件队列
	for {
		var jobEvent *common.JobEvent
		select {
		//监控任务变化事件
		case jobEvent = <-scheduler.jobEventChan:
			//对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-schedulerTimer.C:

		}

		//调度一次任务
		schedulerAfter = scheduler.TryScheduler()
		schedulerTimer.Reset(schedulerAfter)
	}
}

//初始化调度器
//error返回调度器是否初始化成功
func InitScheduler() error {
	G_scheduler = &Scheduler{
		jobEventChan: make(chan *common.JobEvent),
		jobPlanTable: make(map[string]*common.JobSchedulerPlan),
	}

	//启动调度协程
	go G_scheduler.schedulerLoop()

	return nil
}
