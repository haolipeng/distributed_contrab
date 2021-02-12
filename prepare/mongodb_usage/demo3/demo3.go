package main

import (
	"context"
	"fmt"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/clientopt"
	"time"
)

type TimeBeforeCond struct {
	Before int64 `bson:"$lt"`
}

type DeleteCond struct {
	beforeCond TimeBeforeCond `bson:"timePoint.startTime"`
}

//删除记录
func main() {
	var (
		client     *mongo.Client
		err        error
		database   *mongo.Database
		collection *mongo.Collection
		delCond    *DeleteCond
		delResult  *mongo.DeleteResult
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

	delCond = &DeleteCond{
		beforeCond: TimeBeforeCond{Before: time.Now().Unix()},
	}
	delResult, err = collection.DeleteMany(context.TODO(), delCond)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("删除的行数：", delResult.DeletedCount)
}
