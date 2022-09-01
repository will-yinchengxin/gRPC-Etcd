package discovery

import (
	"api-gateway/consts"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc/resolver"
	"strings"
)

type Server struct {
	Name    string `json:"name"`
	Addr    string `json:"addr"`
	Version string `json:"version"`
	Weight  int    `json:"weight"`
}

func BuildPrefix(server Server) string {
	if server.Version == "" {
		return fmt.Sprintf("/%s/", server.Name)
	}
	return fmt.Sprintf("/%s/%s", server.Name, server.Version)
}

func BuildRegisterPath(server Server) string {
	return fmt.Sprintf("%s%s", BuildPrefix(server), server.Addr)
}

func ParseValue(value []byte) (Server, error) {
	srv := Server{}

	if err := json.Unmarshal(value, srv); err != nil {
		return Server{}, err
	}
	return srv, nil
}

func SplitPath(path string) (Server, error) {
	srv := Server{}
	info := strings.Split(path, "/")
	if len(info) == 0 {
		return srv, errors.New("invalid path")
	}

	srv.Addr = info[len(info)-1]
	return srv, nil
}

func Exist(l []resolver.Address, addr resolver.Address) bool {
	for key := range l {
		if l[key].Addr == addr.Addr {
			return true
		}
	}
	return false
}

func Remove(s []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr.Addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}

func BuildResolverUrl(app string) string {
	return consts.Etcd + ":///" + app
}
