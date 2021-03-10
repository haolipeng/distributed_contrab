package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		config            clientv3.Config
		client            *clientv3.Client
		err               error
		kv                clientv3.KV
		putResp           *clientv3.PutResponse
		getResp           *clientv3.GetResponse
		lease             clientv3.Lease
		leaseGrantResp    *clientv3.LeaseGrantResponse
		leaseId           clientv3.LeaseID
		leaseKeepRespChan <-chan *clientv3.LeaseKeepAliveResponse
	)

	//初始化配置，192.168.57.139是etcd的ip地址
	config = clientv3.Config{
		Endpoints:   []string{"192.168.43.185:2379"}, // 集群列表
		DialTimeout: 5 * time.Second,
	}

	// 创建etcd客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	// 创建读写etcd的键值对
	kv = clientv3.NewKV(client)

	//申请一个租约(Lease)
	lease = clientv3.NewLease(client)

	//申请一个10秒的租约
	leaseGrantResp, err = lease.Grant(context.TODO(), 10)

	//获取租约id
	leaseId = leaseGrantResp.ID

	//续约5秒，然后停止续约，10秒生命期 = 15秒生命期
	ctx, _ := context.WithTimeout(context.Background(), time.Second*5)

	//自动续约
	leaseKeepRespChan, err = lease.KeepAlive(ctx, leaseId)

	//自动续约后会将续约结果投递到leaseKeepRespChan通道中
	//开启协程读取通道中信息
	go func() {
		for {
			select {
			case leaseResp := <-leaseKeepRespChan:
				if leaseResp == nil {
					fmt.Println("续约过期了")
					goto END
				} else {
					fmt.Println("收到自动续约:", leaseResp.ID)
				}
			}
		}
	END:
	}()

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
			break
		}

		fmt.Println(time.Now(), getResp.Kvs)
		time.Sleep(2 * time.Second)
	}
}
