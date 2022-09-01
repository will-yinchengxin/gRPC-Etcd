package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Register struct {
	EtcdAddr    []string
	DialTimeOut int

	closeCh     chan struct{}
	leasesID    clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse

	srvInfo Server
	srvTTL  int64
	cli     *clientv3.Client
	logger  *logrus.Logger
}

func NewRegister(etcdAddr []string, logger *logrus.Logger) *Register {
	return &Register{
		EtcdAddr:    etcdAddr,
		DialTimeOut: 3,
		logger:      logger,
	}
}

func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {
	var err error

	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, errors.New("invalid ip addr")
	}

	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddr,
		DialTimeout: time.Duration(r.DialTimeOut) * time.Second,
	}); err != nil {
		return nil, err
	}

	r.srvInfo = srvInfo
	r.srvTTL = ttl

	if err = r.register(); err != nil {
		return nil, err
	}

	r.closeCh = make(chan struct{})

	go r.keepAlive()

	return r.closeCh, nil
}

func (r *Register) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeOut)*time.Second)
	defer cancel()

	/*
		lease 意为租约，类似于 Redis 中的 TTL(Time To Live)
		etcd 中的键值对可以绑定到租约上，实现存活周期控制

		应用客户端可以为 etcd 集群里面的键授予租约
		一旦租约的TTL到期，租约就会过期并且所有附带的键都将被删除
	*/
	leaseResp, err := r.cli.Grant(ctx, r.srvTTL) // 创建一个新的租约
	if err != nil {
		return err
	}

	r.leasesID = leaseResp.ID

	// 试图保持租约永久活着
	if r.keepAliveCh, err = r.cli.KeepAlive(context.Background(), r.leasesID); err != nil {
		return err
	}

	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}

	_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID)) // 节点信息存储至 etcd

	return err
}

func (r *Register) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.srvTTL) * time.Second)

	for {
		select {
		case <-r.closeCh:
			if err := r.unregister(); err != nil {
				r.logger.Error("r.unregister ===> unregister fail, err: ", err)
			}
			// 撤销给定租约
			if _, err := r.cli.Revoke(context.Background(), r.leasesID); err != nil {
				r.logger.Error("r.closeCh ====> revoke fail, err: ", err)
			}
		case res := <-r.keepAliveCh:
			if res == nil {
				if err := r.register(); err != nil {
					r.logger.Error("r.keepAliveCh ====> register fail, err: ", err)
				}
			}
		case <-ticker.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					r.logger.Error("ticker.C === > register fail, err: ", err)
				}
			}
		}
	}
}

func (r *Register) unregister() error {
	// deletes a key, or optionally using WithRange(end), [key, end).
	_, err := r.cli.Delete(context.Background(), BuildRegisterPath(r.srvInfo))
	return err
}

func (r *Register) Stop() {
	r.closeCh <- struct{}{}
}

func (r *Register) UpdateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		weightStr := req.URL.Query().Get("weight")
		weight, err := strconv.Atoi(weightStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		_, _ = w.Write([]byte(err.Error()))

		var update = func() error {
			r.srvInfo.Weight = int64(weight)
			data, err := json.Marshal(r.srvInfo)
			if err != nil {
				return err
			}

			_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
			return err
		}

		if err := update(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		_, _ = w.Write([]byte("update server weight success"))
	}
}

func (r *Register) GetServerInfo() (Server, error) {
	resp, err := r.cli.Get(context.Background(), BuildRegisterPath(r.srvInfo))
	if err != nil {
		return r.srvInfo, err
	}

	server := Server{}
	if resp.Count >= 1 {
		if err := json.Unmarshal(resp.Kvs[0].Value, &server); err != nil {
			return server, err
		}
	}

	return server, err
}
