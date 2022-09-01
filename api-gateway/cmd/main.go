package main

import (
	"api-gateway/cmd/corn"
	"api-gateway/config"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.InitConfig()   // init config
	go corn.StartListen() // 转载路由，将 http 转接成为 gRPC

	{ // 利用信号阻塞
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL)
		s := <-osSignals
		fmt.Println("exit! ", s)
	}
	fmt.Println("gateway listen on :8080")
}
