package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		kv     clientv3.KV
		op     clientv3.Op
		opResp clientv3.OpResponse
	)

	//192.168.57.139是etcd的ip地址，根据自己情况来配置
	config = clientv3.Config{
		Endpoints:   []string{"192.168.43.185:2379"}, // 集群列表
		DialTimeout: 5 * time.Second,
	}

	// 建立一个客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	// 用于读写etcd的键值对
	kv = clientv3.NewKV(client)

	//op put 操作
	op = clientv3.OpPut("/cron/jobs/job3", "haolipeng")
	if opResp, err = kv.Do(context.TODO(), op); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("写入Revision", opResp.Put().Header.Revision)

	//op get 操作
	op = clientv3.OpGet("/cron/jobs/job3")
	if opResp, err = kv.Do(context.TODO(), op); err != nil {
		fmt.Println(err)
		return
	}

	// 打印
	fmt.Println("数据Revision:", opResp.Get().Kvs[0].ModRevision) // create rev == mod rev
	fmt.Println("数据value:", string(opResp.Get().Kvs[0].Value))

	//delete 操作,删除时可返回之前数据
	op = clientv3.OpDelete("/cron/jobs/job3", clientv3.WithPrevKV())
	if opResp, err = kv.Do(context.TODO(), op); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("删除数据Revision:", opResp.Del().Header.Revision)
	fmt.Println("删除数据Value:", opResp.Del().PrevKvs)
}
