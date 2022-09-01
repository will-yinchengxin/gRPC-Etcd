package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"user/config"
	"user/discovery"
	"user/internal/handler"
	"user/internal/repository"
	"user/internal/service"
)

func main() {
	config.InitConfig() // 获取配置信息
	repository.InitDB() // 初始化 DB

	// etcd Addr
	etcdAddress := []string{viper.GetString("etcd.address")}
	// 服务注册
	etcdRegister := discovery.NewRegister(etcdAddress, logrus.New())
	grpcAddr := viper.GetString("server.grpcAddress")
	defer etcdRegister.Stop()

	userNode := discovery.Server{
		Name: viper.GetString("server.domain"),
		Addr: grpcAddr,
	}

	server := grpc.NewServer()
	defer server.Stop()

	// 绑定server
	service.RegisterUserServiceServer(server, handler.NewUserService())
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		panic(err)
	}
	/*
		[root@localhost ~]# etcdctl get --prefix  /user/
		/user/127.0.0.1:10001
		{"name":"user","addr":"127.0.0.1:10001","version":"","weight":0}
	*/
	if _, err = etcdRegister.Register(userNode, 10); err != nil {
		panic(fmt.Sprintf("start server fail: %v", err))
	}

	logrus.Info("server start listen on: ", grpcAddr)
	if err = server.Serve(listener); err != nil {
		panic(err)
	}
}
