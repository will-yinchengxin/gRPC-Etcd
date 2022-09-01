package main

import (
	"context"
	"fmt"
	"proto/proto"
	"google.golang.org/grpc"
	"log"
)
/*
监听server启动的tcp的ip:端口。
新建连接client服务，并绑定tcp。
去调用这2个rpc的函数。
*/
var client proto.UserServerClient

func main() {
	//链接tcp端口
	connect, err := grpc.Dial("127.0.0.1:9527", grpc.WithInsecure())
	if err != nil {
		log.Fatalln(err)
	}
	//新建client
	client = proto.NewUserServerClient(connect)

	//调用
	SaveUser()
	GetUserInfo()
}

func SaveUser() {
	params := proto.UserParams{}
	params.Age = &proto.Age{Age: 31}
	params.Name = &proto.Name{Name: "test"}
	res, err := client.SaveUser(context.Background(), &params)
	if err != nil {
		log.Fatalf("client.SaveUser err: %v", err)
	}
	fmt.Printf("%+v\n", res.Id)
}
func GetUserInfo() {
	res, err := client.GetUserInfo(context.Background(), &proto.Id{Id: 1})
	if err != nil {
		log.Fatalf("client.userInfo err: %v", err)
	}
	fmt.Printf("%+v\n", res)
}