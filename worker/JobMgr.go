package worker

import (
	"github.com/coreos/etcd/clientv3"
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
	var err error
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

	return err
}
