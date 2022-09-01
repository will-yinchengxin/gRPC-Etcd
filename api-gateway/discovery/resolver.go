package discovery

import (
	"context"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
	"time"
)

type Resolver struct {
	schema      string
	EtcdAddr    []string
	DialTimeout int

	closeCh     chan struct{}
	watchCh     clientv3.WatchChan
	cli         *clientv3.Client // etcd 客户端
	keyPrefix   string
	srvAddrList []resolver.Address

	cc     resolver.ClientConn // gRPC 客户端
	logger *logrus.Logger
}

func NewResolver(etcdAddr []string, logger *logrus.Logger) *Resolver {
	return &Resolver{
		EtcdAddr:    etcdAddr,
		logger:      logger,
		DialTimeout: 3,
	}
}

func (r *Resolver) Scheme() string {
	return r.schema
}

func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc

	r.keyPrefix = BuildPrefix(Server{Name: target.Endpoint, Version: target.Authority}) //=>  /name/1.1/ip
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
	var (
		err error
	)
	r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddr,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// 将 解析生成器注入到 解析的 map 中
	// 这个函数只能调用初始化期间(即在一个init()函数),并不是线程安全的。如果注册多个解析器相同的名称,一个注册生效。
	resolver.Register(r)

	r.closeCh = make(chan struct{})

	if err = r.sync(); err != nil { // 同步的从 etcd 中获取信息倒入至 grpc client 中
		return nil, err
	}

	go r.watch()

	return r.closeCh, nil
}

func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrefix, clientv3.WithPrefix())

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
				r.logger.Error("sync failed", err)
			}
		}
	}
}

func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		var info Server
		var err error

		switch ev.Type {
		case clientv3.EventTypePut:
			info, err = ParseValue(ev.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.srvAddrList, addr) {
				r.srvAddrList = append(r.srvAddrList, addr)
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList})
			}
		case clientv3.EventTypeDelete:
			info, err = SplitPath(string(ev.Kv.Key))
			if err != nil {
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

func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.cli.Get(ctx, r.keyPrefix, clientv3.WithPrefix()) // get all endpoint info from etcd
	if err != nil {
		return err
	}
	r.srvAddrList = []resolver.Address{}

	for _, v := range res.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
		r.srvAddrList = append(r.srvAddrList, addr)
	}
	r.cc.UpdateState(resolver.State{Addresses: r.srvAddrList}) // set all the endpoint info into grpc
	return nil
}
