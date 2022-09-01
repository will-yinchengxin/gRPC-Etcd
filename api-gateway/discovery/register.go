package discovery

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Register struct {
	EtcdAddr    []string
	DialTimeout int64

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
		logger:      logger,
		DialTimeout: 5,
	}
}

// Register 注册服务
func (r *Register) Register(srvInfo Server, ttl int64) (chan<- struct{}, error) {
	var (
		err error
	)

	if strings.Split(srvInfo.Addr, ":")[0] == "" {
		return nil, Invalid
	}

	if r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddr,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
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

// StopSrv 停止服务
func (r *Register) StopSrv() {
	r.closeCh <- struct{}{}
}

// GetSrvInfo 更新服务端信息
func (r *Register) GetSrvInfo() (Server, error) {
	var (
		srv Server
	)
	resp, err := r.cli.Get(context.Background(), BuildPrefix(r.srvInfo))
	if err != nil {
		return r.srvInfo, err
	}

	if len(resp.Kvs) >= 1 {
		if err = json.Unmarshal(resp.Kvs[0].Value, &srv); err != nil {
			return r.srvInfo, err
		}
	}

	return srv, nil
}

// UpdateWeightHandler 更新 weight
func (r *Register) UpdateWeightHandler() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		weight, err := strconv.Atoi(request.URL.Query().Get("weight"))
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte{})
			return
		}

		var update = func() error {
			r.srvInfo.Weight = weight
			data, err := json.Marshal(r.srvInfo)
			if err != nil {
				return err
			}

			_, err = r.cli.Put(context.Background(), BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID))
			return err
		}

		if err := update(); err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(err.Error()))
			return
		}

		_, _ = writer.Write([]byte("update server weight success"))
	}
}

func (r *Register) register() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()

	leaseResp, err := r.cli.Grant(ctx, r.srvTTL)
	if err != nil {
		return err
	}

	r.leasesID = leaseResp.ID

	if r.keepAliveCh, err = r.cli.KeepAlive(ctx, r.leasesID); err != nil {
		return err
	}

	data, err := json.Marshal(r.srvInfo)
	if err != nil {
		return err
	}

	if _, err = r.cli.Put(ctx, BuildRegisterPath(r.srvInfo), string(data), clientv3.WithLease(r.leasesID)); err != nil {
		return err
	}
	return nil
}

func (r *Register) unregister() error {
	_, err := r.cli.Delete(context.Background(), BuildPrefix(r.srvInfo))
	return err
}

func (r *Register) revoke() error {
	_, err := r.cli.Revoke(context.Background(), r.leasesID)
	return err
}

func (r *Register) keepAlive() {
	timeTick := time.NewTicker(time.Duration(r.srvTTL))

	for {
		select {
		case <-timeTick.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					r.logger.Log(logrus.PanicLevel, "timeTick.C", err)
				}
			}
		case res := <-r.keepAliveCh:
			if res == nil {
				if err := r.register(); err != nil {
					r.logger.Log(logrus.PanicLevel, "r.keepAliveCh-r.register() ===> ", err)
				}
			}
		case <-r.closeCh:
			if err := r.unregister(); err != nil {
				r.logger.Log(logrus.PanicLevel, "r.closeCh-r.unregister() ===> ", err)
			}

			if err := r.revoke; err != nil {
				r.logger.Log(logrus.PanicLevel, "r.closeCh-r.revoke", err)
			}
		}
	}
}
