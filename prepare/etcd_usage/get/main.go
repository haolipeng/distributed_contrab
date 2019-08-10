package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		client  *clientv3.Client
		config  clientv3.Config
		kv      clientv3.KV
		err     error
		getResp *clientv3.GetResponse
	)

	//初始化配置
	config = clientv3.Config{
		Endpoints:   []string{"192.168.57.139:2379"},
		DialTimeout: 5 * time.Second,
	}

	//初始化客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println("init etcd client error")
		return
	}

	//创建用于键值对的KV
	kv = clientv3.NewKV(client)

	//提前写入两条数据，用于查询
	_, err = kv.Put(context.TODO(), "/cron/jobs/job1", "haolipeng")
	_, err = kv.Put(context.TODO(), "/cron/jobs/job2", "zhouyang")

	//get + WithPrefix 前缀匹配获取
	getResp, err = kv.Get(context.TODO(), "/cron/jobs/", clientv3.WithPrefix())
	if err != nil {
		fmt.Println("etcd get opertion failed!")
		return
	} else {
		fmt.Printf("key:%s, value:%s\n", getResp.Kvs[0].Key, getResp.Kvs[0].Value)
		fmt.Printf("key:%s, value:%s\n", getResp.Kvs[1].Key, getResp.Kvs[1].Value)
	}

	//get + 只获取个数,此时Kvs数组中没有值
	//更多的OpOption可参考官网手册
	getResp, err = kv.Get(context.TODO(), "/cron/jobs/", clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		fmt.Println("etcd get opertion failed!")
		return
	} else {
		fmt.Println(getResp.Kvs, getResp.Count)
	}
}
