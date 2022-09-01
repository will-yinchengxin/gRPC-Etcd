package main

//采用https的token加密

import (
  "context"
  "fmt"
  "proto/proto"
  "google.golang.org/grpc"
  "google.golang.org/grpc/credentials"
  "log"
  "math/rand"
  "net"
)

type UserServer struct{}

func main() {

  //读2个证书
  c, err := credentials.NewServerTLSFromFile("/root/goTest/openssl/proto.pem", "/root/goTest/openssl/proto.key")
  if err != nil {
    log.Fatalf("new tls server err:", err.Error())
  }

  //监听端口
  listen, err := net.Listen("tcp", "127.0.0.1:9528")
  if err != nil {
    log.Fatalf("tcp listen failed:%v", err)
  }

  //新建gRPC服务，并且传入证书handle
  server := grpc.NewServer(grpc.Creds(c))

  fmt.Println("userServer grpc services start success")

  //注册本次的UserServer 服务
  proto.RegisterUserServerServer(server, &UserServer{})
  _ = server.Serve(listen)
}

//保存用户
func (Service *UserServer) SaveUser(ctx context.Context, params *proto.UserParams) (*proto.Id, error) {
  id := rand.Int31n(100) //随机生成id 模式保存成功
  res := &proto.Id{Id: id}
  fmt.Printf("%+v ", params.GetAge())
  fmt.Printf("%+v\n", params.GetName())
  return res, nil
}

func (Service *UserServer) GetUserInfo(ctx context.Context, id *proto.Id) (*proto.UserInfo, error) {
  res := &proto.UserInfo{Id: id.GetId(), Name: "test", Age: 31}
  return res, nil
}
