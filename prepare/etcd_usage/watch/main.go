package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"time"
)

func main() {
	var (
		config             clientv3.Config
		client             *clientv3.Client
		err                error
		kv                 clientv3.KV
		watchRespChan      clientv3.WatchChan
		watchResp          clientv3.WatchResponse
		event              *clientv3.Event
		putResp            *clientv3.PutResponse
		deleteResp         *clientv3.DeleteResponse
		getResp            *clientv3.GetResponse
		watchStartRevision int64
		watcher            clientv3.Watcher
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

	//关闭资源
	defer client.Close()

	//创建读写键值对的KV
	kv = clientv3.NewKV(client)

	//模拟etcd中的kv变化
	go func() {
		for {
			putResp, err = kv.Put(context.TODO(), "/cron/jobs/job1", "123")
			deleteResp, err = kv.Delete(context.TODO(), "/cron/jobs/job1")

			time.Sleep(time.Second * 1)
		}
	}()

	//先GET到当前值并监听后续变化
	if getResp, err = kv.Get(context.TODO(), "/cron/jobs/job1"); err != nil {
		fmt.Println(err)
		return
	}

	//当前etcd集群的事务id，此值是单调递增的
	watchStartRevision = getResp.Header.Revision + 1

	//创建一个Watcher
	watcher = clientv3.NewWatcher(client)

	//启动监听
	fmt.Println("从该版本向后监听：", watchStartRevision)

	ctx, cancelFun := context.WithCancel(context.TODO())
	time.AfterFunc(5*time.Second, func() {
		cancelFun()
	})

	watchRespChan = watcher.Watch(ctx, "/cron/jobs/job1", clientv3.WithRev(watchStartRevision))

	// 处理kv变化事件
	for watchResp = range watchRespChan {
		for _, event = range watchResp.Events {
			switch event.Type {
			case mvccpb.PUT:
				fmt.Println("修改为:", string(event.Kv.Value), "Revision:", event.Kv.CreateRevision, event.Kv.ModRevision)
			case mvccpb.DELETE:
				fmt.Println("删除了", "Revision:", event.Kv.ModRevision)
			}
		}
	}

	////normal watch
	//watchRespChan = client.Watch(context.TODO(), "/cron/jobs/", clientv3.WithPrefix())
	////watch有几种错误
	//for wresp :=range watchRespChan{
	//	wresp.Events
	//	for _,ev := range  wresp.Events{
	//		fmt.Printf("%s %q : %q\n", ev.Type, ev.Kv.Key, ev.Kv.Value)
	//	}
	//}
}
