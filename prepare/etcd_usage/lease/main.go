package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		config         clientv3.Config
		client         *clientv3.Client
		err            error
		kv             clientv3.KV
		putResp        *clientv3.PutResponse
		getResp        *clientv3.GetResponse
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
	)

	//初始化配置，192.168.57.139是etcd的ip地址
	config = clientv3.Config{
		Endpoints:   []string{"192.168.57.139:2379"}, // 集群列表
		DialTimeout: 5 * time.Second,
	}

	// 建立一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	// 用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	//申请一个租约(Lease)
	lease = clientv3.NewLease(client)

	//申请一个十秒的租约
	leaseGrantResp, err = lease.Grant(context.TODO(), 10)

	//获取租约id
	leaseId = leaseGrantResp.ID

	//put一个kv，让它和租约关联，实现10秒后自动过期
	if putResp, err = kv.Put(context.TODO(), "/cron/lock/job1", "lock1", clientv3.WithLease(leaseId)); err != nil {
		fmt.Println(err)
	}

	fmt.Println(putResp.Header.Revision)

	//不断检查下值是否过期
	for {
		//get操作来获取值
		getResp, err = kv.Get(context.TODO(), "/cron/lock/job1")
		if err != nil {
			break
		}
		if getResp.Count == 0 {
			fmt.Println("已经过期了")
		}

		fmt.Println(time.Now(), getResp.Kvs)
		time.Sleep(2 * time.Second)
	}
}
