package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
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

// jobName过滤条件
type FindByJobName struct {
	JobName string `bson:"jobName"` // JobName赋值为job10
}

func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
		result     *mongo.InsertOneResult
		cursor     mongo.Cursor
		record     *LogRecord
		cond       *FindByJobName
	)
	//1、连接数据库
	client, err = mongo.Connect(context.TODO(), "mongodb://127.0.0.1:27017", clientopt.ConnectTimeout(5*time.Second))
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
	fmt.Println("insertOne success,insertedID:", docId.Hex())

	//cond := &LogRecord{JobName: "Job10"}
	// 4, 按照jobName字段过滤, 想找出jobName=job10,
	cond = &FindByJobName{JobName: "Job10"} // {"jobName": "job10"}
	cnt, err := collection.Count(context.TODO(), cond)
	if err != nil {
		fmt.Println("collection Find failed")
		return
	}
	fmt.Printf("collection count:%d\n", cnt)

	//5, 按照jobName字段过滤, 想找出jobName=job10,限制只寻找五条
	cursor, err = collection.Find(context.TODO(), cond, findopt.Skip(0), findopt.Limit(5))
	if err != nil {
		fmt.Println("collection Find failed")
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		//定义一个日志对象
		r := &LogRecord{}

		//反序列化bson到对象
		if err = cursor.Decode(r); err != nil {
			fmt.Println(err)
			return
		}

		//将日志打印出来
		fmt.Println(*r)
	}
}
