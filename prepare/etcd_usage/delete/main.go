package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		config     clientv3.Config
		client     *clientv3.Client
		err        error
		kv         clientv3.KV
		deleteResp *clientv3.DeleteResponse
	)

	//192.168.57.139是etcd的ip地址，根据自己情况来配置
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

	//添加两个元素，测试前缀匹配删除
	_, err = kv.Put(context.TODO(), "/cron/jobs/task1", "1")
	_, err = kv.Put(context.TODO(), "/cron/jobs/task2", "2")

	//删除操作,删除之前先put一项，key:/cron/jobs/job1 value:bye
	if deleteResp, err = kv.Delete(context.TODO(), "/cron/jobs/", clientv3.WithPrevKV(), clientv3.WithPrefix()); err != nil {
		fmt.Println(err)
		return
	}

	//被删除之前的key是什么
	if len(deleteResp.PrevKvs) > 0 {
		for _, kvPair := range deleteResp.PrevKvs {
			fmt.Println("删除了", string(kvPair.Key), string(kvPair.Value))
		}
	}
}
