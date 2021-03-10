package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"time"
)

func main() {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		err     error
		leaseId clientv3.LeaseID
		lease   clientv3.Lease

		leaseGrantResp *clientv3.LeaseGrantResponse
		/*kv					clientv3.KV
		putResp				*clientv3.PutResponse
		getResp				*clientv3.GetResponse*/
		keepAliveRespChan <-chan *clientv3.LeaseKeepAliveResponse
		keepAliveResp     *clientv3.LeaseKeepAliveResponse
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

	//创建一个租约(Lease)
	lease = clientv3.NewLease(client)

	// 申请一个5s的租约，观察其是否过期
	leaseGrantResp, err = lease.Grant(context.TODO(), 10)
	if err != nil {
		fmt.Println(err)
		return
	}

	//获取租约id
	leaseId = leaseGrantResp.ID

	//创建context上下文，5秒后超市
	ctx, _ := context.WithTimeout(context.TODO(), 3*time.Second)

	// 自动续租
	keepAliveRespChan, err = lease.KeepAlive(ctx, leaseId)
	if err != nil {
		fmt.Println(err)
		return
	}

	//监视续租的变化情况
	go func() {
		for {
			select {
			case keepAliveResp = <-keepAliveRespChan:
				//以前的示例代码有问题
				if keepAliveResp == nil {
					fmt.Println("租约已经失效!", time.Now().String())
					goto END
				} else if keepAliveResp != nil {
					//每秒会续租一次，所以会受到一次续租应答
					fmt.Println("收到续租应答:", keepAliveResp.ID, time.Now().String())
				}
			}
		}
	END:
	}()

	/*// 创建KV
	kv = clientv3.NewKV(client)

	// put操作，将数据压入到
	putKey := "/cron/jobs/job1"
	putVal := ""
	putResp,err = kv.Put(context.TODO(),putKey,putVal,clientv3.WithLease(leaseId))
	if err != nil{
		fmt.Println(err,putResp.PrevKv)
		return
	}

	// 监听5s租约的
	go func() {
		for {
			getResp,err = kv.Get(context.TODO(),putKey)
			if err != nil{
				fmt.Print(err)
				break
			}

			//判断是否过期
			if getResp.Count == 0{
				fmt.Println("已经过期了")
				break
			}

			fmt.Println("还没过期:", getResp.Kvs)
			time.Sleep(2 * time.Second)
		}
	}()*/

	//手动睡眠30秒
	time.Sleep(30 * time.Second)
	fmt.Println("the pro")
}
