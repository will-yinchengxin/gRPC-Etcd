package main

import (
	"context"
	"fmt"
	"proto/proto"
	"google.golang.org/grpc"
	"log"
	"math/rand"
	"net"
)

/*
新建tcp连接。
注册gRPC服务，并把它挂到tcp上。
完成对外提供的几个rpc方法的逻辑
*/

type UserServer struct{}

func main() {

	//监听tcp
	listen, err := net.Listen("tcp", "127.0.0.1:9527")
	if err != nil {
		log.Fatalf("tcp listen failed:%v", err)
	}

	//新建gRPC
	server := grpc.NewServer()
	fmt.Println("userServer grpc services start success")

	//rcp方法注册到grpc
	proto.RegisterUserServerServer(server, &UserServer{})

	//监听tcp
	_ = server.Serve(listen)
}

//保存用户
//第一个参数是固定的context
func (Service *UserServer) SaveUser(ctx context.Context, params *proto.UserParams) (*proto.Id, error) {
	id := rand.Int31n(100) //随机生成id 模式保存成功
	res := &proto.Id{Id: id}
	fmt.Printf("%+v", params.GetAge())
	fmt.Printf("%+v\n", params.GetName())
	return res, nil
}

//获取用户信息
//第一个参数是固定的context
func (Service *UserServer) GetUserInfo(ctx context.Context, id *proto.Id) (*proto.UserInfo, error) {
	res := &proto.UserInfo{Id: id.GetId(), Name: "test", Age: 31}
	return res, nil
}
