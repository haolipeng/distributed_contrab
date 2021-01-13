package worker

import (
	"github.com/coreos/etcd/clientv3"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

//任务管理的单例
var (
	G_JobMgr *JobMgr
)

//初始化任务管理器,或者返回JobMrg指针变量，和etcd建立连接
func InitJobMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
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

	//新建用于读写键值对的kv
	kv = clientv3.NewKV(client)

	//新建租约
	lease = clientv3.NewLease(client)

	//赋值单例
	G_JobMgr = &JobMgr{
		client: client,
		lease:  lease,
		kv:     kv,
	}

	return err
}
