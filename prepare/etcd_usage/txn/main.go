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
		lease          clientv3.Lease
		leaseGrantResp *clientv3.LeaseGrantResponse
		leaseId        clientv3.LeaseID
		keepRespChan   <-chan *clientv3.LeaseKeepAliveResponse
		keepResp       *clientv3.LeaseKeepAliveResponse
		ctx            context.Context
		cancelFunc     context.CancelFunc
		kv             clientv3.KV
		txn            clientv3.Txn
		txnResp        *clientv3.TxnResponse
	)

	//配置信息 Endpoints 是集群服务器地址
	config = clientv3.Config{
		Endpoints:   []string{"192.168.43.185:2379"},
		DialTimeout: 5 * time.Second,
	}

	//创建客户端
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	//lease实现锁自动过期
	//op操作
	//txn事务：if else then
	//创建租约

	//一.上锁
	//(创建租约、自动续租、拿着租约去抢占锁，抢占期间要确保租约是有效的)
	lease = clientv3.NewLease(client)

	// 申请一个5s的租约，观察其是否过期
	leaseGrantResp, err = lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Println(err)
		return
	}

	//获取租约id
	leaseId = leaseGrantResp.ID

	//准备一个取消自动续租的context
	ctx, cancelFunc = context.WithCancel(context.TODO())

	//确保函数退出后，自动续租会停止
	defer cancelFunc()
	defer lease.Revoke(context.TODO(), leaseId) //释放租约

	// 自动续租
	keepRespChan, err = lease.KeepAlive(ctx, leaseId)
	if err != nil {
		fmt.Println(err)
		return
	}

	//监视续租的变化情况
	go func() {
		for {
			select {
			case keepResp = <-keepRespChan:
				//以前的示例代码有问题
				if keepResp == nil {
					fmt.Println("租约已经失效!", time.Now().String())
					goto END
				} else if keepRespChan != nil {
					//每秒会续租一次，所以会受到一次续租应答
					fmt.Println("收到续租应答:", keepResp.ID, time.Now().String())
				}
			}
		}
	END:
	}()

	//二.业务逻辑操作
	//0.创建KV键
	kv = clientv3.NewKV(client)

	//1.创建事务
	txn = kv.Txn(context.TODO())

	//2.定义事务
	//如果key不存在
	txn.If(clientv3.Compare(clientv3.CreateRevision("/cron/jobs/job9"), "=", 0)).
		Then(clientv3.OpPut("/cron/jobs/job9", "xxx", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet("/cron/jobs/job9"))

	//3.提交事务
	if txnResp, err = txn.Commit(); err != nil {
		fmt.Println(err)
		return
	}

	//4.判断是否成功抢到锁
	if !txnResp.Succeeded {
		fmt.Println("锁被占用:", string(txnResp.Responses[0].GetResponseRange().Kvs[0].Value))
		return
	} else {
		fmt.Println("成功抢占到锁")
	}

	//处理业务逻辑
	fmt.Println("处理任务")
	time.Sleep(5 * time.Second)

	//三.释放锁
	//前面已经完成此操作
	//defer cancelFunc()
	//defer lease.Revoke(context.TODO(),leaseId) //释放租约
}
