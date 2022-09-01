package corn

import (
	"api-gateway/discovery"
	"api-gateway/internal/service"
	"api-gateway/middleware/wrapper"
	"api-gateway/pkg/rpc"
	"api-gateway/pkg/util"
	"api-gateway/routes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc/resolver"
	"net/http"
	"time"
)

func StartListen() {
	// init etcd
	etcdAddr := []string{viper.GetString("etcd.address")}
	etcdRegister := discovery.NewResolver(etcdAddr, logrus.New())
	// etcd 注册
	resolver.Register(etcdRegister)

	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)

	// get server name
	userServerName := viper.GetString("domain.user")

	// build rpc conn
	userConn, err := rpc.RPCConnect(ctx, userServerName, etcdRegister)
	if err != nil {
		return
	}
	userService := service.NewUserServiceClient(userConn)

	// add user server into wrapper
	wrapper.NewServiceWrapper(userServerName)

	ginRouter := routes.NewRouter(userService)
	server := &http.Server{
		Addr:           viper.GetString("server.port"),
		Handler:        ginRouter,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err = server.ListenAndServe()

	if err != nil {
		fmt.Println("server.ListenAndServe() fail: ", err, " will do gracefully shut down")
	}

	go func() {
		util.GracefullyShutdown(server)
	}()

	if err = server.ListenAndServe(); err != nil {
		fmt.Println("server.ListenAndServe()", err)
	}
}
