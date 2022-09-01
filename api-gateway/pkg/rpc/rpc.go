package rpc

import (
	"api-gateway/discovery"
	"context"
	"fmt"
	"google.golang.org/grpc"
)

func RPCConnect(ctx context.Context, serviceName string, etcdRegister *discovery.Resolver) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	target := fmt.Sprintf("%s:///%s", etcdRegister.Scheme(), serviceName)
	conn, err = grpc.DialContext(ctx, target, opts...)
	return
}
