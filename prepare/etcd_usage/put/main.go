package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

//func main() {
//	var (
//		config clientv3.Config
//		kv clientv3.KV
//		client *clientv3.Client
//		err error
//		putResp *clientv3.PutResponse
//	)
//
//	//初始化配置信息
//	config = clientv3.Config{
//		Endpoints: []string{"192.168.57.139"},
//		DialTimeout: 5 * time.Second,
//	}
//
//	//创建etcd客户端
//	if client,err = clientv3.New(config); err != nil {
//	}
//
//	//用于读写etcd的键值对
//	kv = clientv3.NewKV(client)
//	_, cancelFun := context.WithTimeout(context.Background(),time.Second * 5)
//	putResp,err = kv.Put(context.Background(), "/cron/jobs/job1", "hello world")
//	cancelFun()
//
//	if err != nil{
//		fmt.Println(err)
//	}else {
//		fmt.Println("revision", putResp.Header.Revision)
//		if putResp.PrevKv != nil{
//			fmt.Println("preValue:", string(putResp.PrevKv.Value))
//		}
//	}
//
//	return
//}

func main() {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		err     error
		kv      clientv3.KV
		putResp *clientv3.PutResponse
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

	//put操作
	//实验证明，即使是更新相同值，也是会改变revision值
	if putResp, err = kv.Put(context.TODO(), "/cron/jobs/job1", "bye", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Revision:", putResp.Header.Revision)
		if putResp.PrevKv != nil { // 打印hello
			fmt.Println("PrevValue:", string(putResp.PrevKv.Value))
		}
	}
}
