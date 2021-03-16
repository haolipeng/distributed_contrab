package worker

import (
	"context"
	"distributed_contrab/common"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

type LogSink struct {
	client         *mongo.Client
	logCollection  *mongo.Collection
	logChan        chan *common.JobLog
	autoCommitChan chan *common.LogBatch
}

var (
	G_logSink *LogSink
)

func InitLogSink() error {
	//建立mongodb连接
	client, err := mongo.Connect(context.TODO(), G_config.MongodbUri, clientopt.ConnectTimeout(time.Duration(G_config.MongodbConnectTimeout)*time.Millisecond))
	if err != nil {
		return err
	}

	//选择db和collection
	G_logSink = &LogSink{
		client:         client,
		logCollection:  client.Database("cron").Collection("log"),
		logChan:        make(chan *common.JobLog, 1000),
		autoCommitChan: make(chan *common.LogBatch, 1000),
	}

	// 启动一个mongodb 处理协程
	go G_logSink.writeLoop()
	return nil
}

// 批量写入日志
func (logSink *LogSink) saveLogs(batch *common.LogBatch) {
	_, err := logSink.logCollection.InsertMany(context.TODO(), batch.Logs)
	if err != nil {
		fmt.Println("LogSink saveLogs function error:", err)
	}
}

func (logSink *LogSink) writeLoop() {
	var (
		log          *common.JobLog
		logBatch     *common.LogBatch // 当前的批次
		commitTimer  *time.Timer
		timeoutBatch *common.LogBatch
	)

	for {
		select {
		case log = <-logSink.logChan:
			if logBatch == nil {
				logBatch = &common.LogBatch{}

				//让这个批次超时自动提交
				commitTimer = time.AfterFunc(
					time.Duration(G_config.JobLogCommitTimeout)*time.Millisecond,
					func(batch *common.LogBatch) func() {
						return func() {
							logSink.autoCommitChan <- batch
						}
					}(logBatch),
				)
			}

			//将新日志追加到批次中
			logBatch.Logs = append(logBatch.Logs, log)

			//添加调试日志
			fmt.Println("log = <-logSink.logChan case is in!")

			//如果批次满了，则触发提交操作
			if len(logBatch.Logs) >= G_config.JobLogBatchSize {
				//发送日志
				logSink.saveLogs(logBatch)
				//清空logBathc
				logBatch = nil
				//停止定时器
				commitTimer.Stop()
			}
		case timeoutBatch = <-logSink.autoCommitChan: //过期的批次
			//判断过期批次是否仍旧是当前的批次
			if timeoutBatch != logBatch {
				continue
			}

			// 将批次写入mongodb中
			logSink.saveLogs(timeoutBatch)

			//清空logBatch
			logBatch = nil
		}
	}
}

//收集任务执行日志
func (logSink *LogSink) Append(jobLog *common.JobLog) {
	select {
	case logSink.logChan <- jobLog:
	default:
	}
}
