package discovery

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"time"
)

const (
	scheme = "etcd"
)

// Resolver 解析 grpc 客户端
type Resolver struct {
	schema      string
	EtcdAddr    []string
	DialTimeOut int

	closeCh     chan struct{}
	watchCh     clientv3.WatchChan
	cli         *clientv3.Client // 提供管理etcd的客户端会话
	keyPrefix   string
	srvAddrList []resolver.Address // 代表一个客户端连接到服务器

	cc     resolver.ClientConn // 回调通知任务更新 grpc 客户端
	logger *logrus.Logger      // 日志
}

// NewResolver Build Base on etcd
func NewResolver(etcdAddr []string, logger *logrus.Logger) *Resolver {
	return &Resolver{
		schema:      scheme,
		EtcdAddr:    etcdAddr,
		DialTimeOut: 3,
		logger:      logger,
	}
}

func (r *Resolver) Scheme() string {
	return r.schema
}

func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc

	r.keyPrefix = BuildPrefix(Server{Name: target.Endpoint, Version: target.Authority})
	if _, err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

func (r *Resolver) Close() {
	r.closeCh <- struct{}{}
}

func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	r.cli, err = clientv3.New(clientv3.Config{ // 实例化 etcd 客户端
		Endpoints:   r.EtcdAddr,
		DialTimeout: time.Duration(r.DialTimeOut) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resolver.Register(r)

	r.closeCh = make(chan struct{})

	if err = r.sync(); err != nil {
		return nil, err
	}
	go r.watch()

	return r.closeCh, nil
}

// 同步获取所有地址信息
func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.cli.Get(ctx, r.keyPrefix, clientv3.WithPrefix()) // etcd 总 get res with prefix 方法
	if err != nil {
		return err
	}

	r.srvAddrList = []resolver.Address{}
	for key, val := range res.Kvs {
		fmt.Println(key, val)
		info, err := ParseVal(val.Value)
		if err != nil {
			fmt.Println(err)
			continue
		}
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight} // grpc 地址
		r.srvAddrList = append(r.srvAddrList, addr)
	}
	r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList}) // 更新客户端连接，grpc
	return nil
}

func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix()) // etcd 定时查看客户端连接

	for {
		select {
		case <-r.closeCh:
			return
		case res, ok := <-r.watchCh:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				r.logger.Error("sync fail", err)
			}
		}
	}
}

func (r *Resolver) update(events []*clientv3.Event) {
	for key, val := range events {
		var (
			info Server
			err  error
		)
		fmt.Println("print key val", key, val)

		switch val.Type {
		case clientv3.EventTypePut:
			info, err = ParseVal(val.Kv.Value)
			if err != nil {
				fmt.Println(err)
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.srvAddrList, addr) {
				r.srvAddrList = append(r.srvAddrList, addr)
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList})
			}
		case clientv3.EventTypeDelete:
			info, err = SplitPath(string(val.Kv.Key))
			if err != nil {
				fmt.Println(err)
				continue
			}
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrList, addr); ok {
				r.srvAddrList = s
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList})
			}
		}
	}
}
