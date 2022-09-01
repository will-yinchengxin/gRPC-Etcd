package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"proto/proto"

	"log"
)

var client proto.UserServerClient

func main() {

	//读取证书和服务名
	crt, err := credentials.NewClientTLSFromFile("/root/goTest/openssl/proto.pem", "www.zchd.ltd")
	if err != nil {
		panic(err.Error())
	}

	//监听端口，并传入证书handle
	connect, err := grpc.Dial("127.0.0.1:9528", grpc.WithTransportCredentials(crt))
	if err != nil {
		log.Fatalln(err)
	}

	defer connect.Close()

	//新建服务客户端
	client = proto.NewUserServerClient(connect)

	SaveUser()
	//GetUserInfo()
}

func SaveUser() {
	params := proto.UserParams{}
	params.Age = &proto.Age{Age: 31}
	params.Name = &proto.Name{Name: "test"}
	res, err := client.SaveUser(context.Background(), &params)
	if err != nil {
		log.Fatalf("client.SaveUser err: %v", err)
	}
	fmt.Println(res.Id)
}
func GetUserInfo() {
	res, err := client.GetUserInfo(context.Background(), &proto.Id{Id: 1})
	if err != nil {
		log.Fatalf("client.userInfo err: %v", err)
	}
	fmt.Printf("%+v\n", res)
}
