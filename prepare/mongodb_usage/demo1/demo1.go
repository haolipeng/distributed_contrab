package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

type TimePoint struct {
	StartTime int64 `bson:"startTime"`
	EndTime   int64 `bson:"endTime"`
}

type LogRecord struct {
	JobName   string    `bson:"jobName"`
	Command   string    `bson:"command"`
	Err       string    `bson:"err"`
	Content   string    `bson:"content"`
	TimePoint TimePoint `bson:"timePoint"` // 执行时间点
}

func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
		result     *mongo.InsertOneResult
		manyResult *mongo.InsertManyResult
		record     *LogRecord
	)
	//1、连接数据库
	client, err = mongo.Connect(context.TODO(), "mongodb://192.168.101.240:27017", clientopt.ConnectTimeout(5*time.Second))
	if err != nil {
		fmt.Println("Connect mongodb failed,err:", err)
		return
	}

	//2、选择数据库
	database = client.Database("my_db")

	//3、选择Collection集合
	collection = database.Collection("my_collection")

	//4、构造数据记录
	record = &LogRecord{
		JobName:   "Job10",
		Command:   "echo helloworld",
		Err:       "",
		Content:   "hello",
		TimePoint: TimePoint{StartTime: time.Now().Unix(), EndTime: time.Now().Unix() + 10},
	}

	//5、插入元素
	if result, err = collection.InsertOne(context.TODO(), record); err != nil {
		fmt.Println("collection insert failed!")
		return
	}

	//6、插入是否成功
	docId := result.InsertedID.(objectid.ObjectID)
	fmt.Println("自增ID:", docId.Hex())

	//7、批量插入多条记录
	logRecordArray := []interface{}{
		record,
		record,
		record,
	}
	//api的第二个参数是[]interface{}，interface类型值的切片
	manyResult, err = collection.InsertMany(context.TODO(), logRecordArray)
	if err != nil {
		fmt.Println(err)
		return
	}

	//8、查看所有插入的id索引
	for _, insertId := range manyResult.InsertedIDs {
		// 拿着interface{}， 反射成objectID
		docId = insertId.(objectid.ObjectID)
		fmt.Println("自增ID:", docId.Hex())
	}
}
