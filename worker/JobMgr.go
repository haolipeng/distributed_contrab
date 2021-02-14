package worker

import (
	"context"
	"distributed_contrab/common"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

type JobMgr struct {
	client  *clientv3.Client
	kv      clientv3.KV
	lease   clientv3.Lease
	watcher clientv3.Watcher
}

//任务管理的单例
var (
	G_JobMgr *JobMgr
)

//监听任务变化
func (jobMgr *JobMgr) watchJobs() error {
	var (
		err     error
		getResp *clientv3.GetResponse
	)

	//1.获取/cron/jobs目录下所有任务
	getResp, err = jobMgr.client.Get(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithPrefix())
	if err != nil {
		return err
	}

	//2.将已存在的任务投递到调度器中
	for _, kvpair := range getResp.Kvs {
		var job *common.Job
		job, err = common.UnpackJob(kvpair.Value)
		if err != nil {

		}
		common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
		//TODO:将job事件投递给调度器
	}

	//3.监听所有任务的变化
	go func() {
		watchStartVersion := getResp.Header.Revision + 1
		//监听cron/jobs目录的后续变化
		watchChan := jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_DIR, clientv3.WithRev(watchStartVersion))
		for watchResp := range watchChan {
			for _, event := range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT:
					//生成job删除事件，
					var job *common.Job
					if job, err = common.UnpackJob(event.Kv.Value); err != nil {
						continue
					}
					common.BuildJobEvent(common.JOB_EVENT_SAVE, job)
					//TODO:将job事件投递给调度器
				case mvccpb.DELETE:
					var job *common.Job
					if job, err = common.UnpackJob(event.Kv.Value); err != nil {
						continue
					}
					common.BuildJobEvent(common.JOB_EVENT_DELETE, job)
					//TODO:将job事件投递给调度器
				}
			}
		}
	}()

	//put变化
	//delete变化
	//TODO:not implement
	return err
}

//监听强杀任务通知
func (jobMgr *JobMgr) watchKiller() error {
	var err error
	//TODO:not implement
	return err
}

//初始化任务管理器,或者返回JobMrg指针变量，和etcd建立连接
func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   G_config.EtcdEndpoints,                                     //集群地址
		DialTimeout: time.Duration(G_config.EtcdDialTimeout) * time.Millisecond, //连接超时
	}

	//新建连接
	if client, err = clientv3.New(config); err != nil {
		return err
	}

	//得到KV和Lease的API子集
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	//赋值单例
	G_JobMgr = &JobMgr{
		client:  client,
		lease:   lease,
		kv:      kv,
		watcher: watcher,
	}

	//启动任务监听
	go G_JobMgr.watchJobs()

	return err
}
