package master

import (
	"context"
	"distributed_contrab/common"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
	"time"
)

type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_logMgr *LogMgr
)

func (logMgr *LogMgr) ListLog(name string, skip int, limit int) ([]*common.JobLog, error) {
	// 过滤条件
	logFilter := common.JobLogFilter{JobName: name}

	//按照任务开始时间倒排
	logSort := common.SortLogByStartTime{SortOrder: -1}

	logArrJobs := make([]*common.JobLog, 0)

	//开始查询
	cursor, err := logMgr.logCollection.Find(
		context.TODO(),
		logFilter,
		findopt.Sort(logSort),
		findopt.Skip(int64(skip)),
		findopt.Limit(int64(limit)),
	)
	if err != nil {
		return logArrJobs, err
	}

	//不要忘记释放的事
	if cursor != nil {
		defer cursor.Close(context.TODO())
	}

	for cursor.Next(context.TODO()) {
		var jobLog *common.JobLog
		//反序列化
		err = cursor.Decode(jobLog)
		if err != nil {
			//日志不合法，继续
			continue
		}

		//解析后的日志压入logArrJobs中
		logArrJobs = append(logArrJobs, jobLog)
	}

	return logArrJobs, nil
}

func InitLogMgr() error {
	var (
		err error
	)

	//1. 建立mongodb连接
	client, err := mongo.Connect(
		context.TODO(),
		G_config.MongodbUri,
		clientopt.ConnectTimeout(time.Duration(G_config.MongodbConnectTimeout)*time.Millisecond))
	if err != nil {
		fmt.Println("Connect mongodb failed,err:", err)
		return err
	}

	collection := client.Database("cron").Collection("log")
	//2. 返回构建的对象
	G_logMgr = &LogMgr{
		client:        client,
		logCollection: collection,
	}

	return err
}
