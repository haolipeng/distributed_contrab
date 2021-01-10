package master

import (
	"context"
	"distributed_contrab/common"
	"encoding/json"
	"github.com/coreos/etcd/clientv3"
	"time"
)

type JobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

//单例对象 用于任务管理
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

//保存任务，保存成功返回上一个任务
func (jobMrg *JobMgr) JobSave(job *common.Job) (oldJob *common.Job, err error) {
	var (
		jobKey   string
		jobValue []byte
		putResp  *clientv3.PutResponse
		LastJob  common.Job
	)

	jobKey = common.JOB_SAVE_DIR + job.Name

	//1.将任务信息json序列化
	if jobValue, err = json.Marshal(job); err != nil {
		return
	}

	//返回上一次的job
	if putResp, err = jobMrg.kv.Put(context.TODO(), jobKey, string(jobValue), clientv3.WithPrevKV()); err != nil {
		return
	}

	if putResp.PrevKv != nil {
		//将返回值的value部分进行反序列化
		if err = json.Unmarshal(putResp.PrevKv.Value, &LastJob); err == nil {
			oldJob = &LastJob
			err = nil
		}
	}

	return
}

//删除任务，并返回上一次任务信息
func (jobMgr *JobMgr) JobDelete(name string) (oldJob *common.Job, err error) {
	var (
		jobKey     string
		deleteResp *clientv3.DeleteResponse
		lastJob    common.Job
	)

	//删除指定任务
	jobKey = common.JOB_SAVE_DIR + name
	if deleteResp, err = jobMgr.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}

	if len(deleteResp.PrevKvs) > 0 {
		if err = json.Unmarshal(deleteResp.PrevKvs[0].Value, &lastJob); err != nil {
			err = nil
			return
		}

		oldJob = &lastJob
	}

	return
}

//枚举所有任务信息
func (jobMrg *JobMgr) JobList() ([]*common.Job, error) {
	var (
		dirName string
		err     error
		getResp *clientv3.GetResponse
		jobList []*common.Job
		job     *common.Job //临时存储job的结构体
		//kvPairs *mvccpb.KeyValue
	)

	//申请切片
	jobList = make([]*common.Job, 0)

	//获取目录下所有任务信息
	dirName = common.JOB_SAVE_DIR
	if getResp, err = jobMrg.kv.Get(context.TODO(), dirName, clientv3.WithPrefix()); err != nil {
		goto ERR
	}

	if len(getResp.Kvs) > 0 {
		for _, kvPairs := range getResp.Kvs {
			job = &common.Job{}
			//反序列化json
			if err = json.Unmarshal(kvPairs.Value, job); err != nil {
				err = nil
				continue
			}

			//将任务添加到任务列表中
			jobList = append(jobList, job)
		}
	}

	return jobList, err

ERR:
	return jobList, err
}
