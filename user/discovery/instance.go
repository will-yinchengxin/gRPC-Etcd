package discovery

import (
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
	Weight  int64  `json:"weight"`
}

func BuildPrefix(srv Server) string {
	if srv.Version == "" {
		return fmt.Sprintf("/%s/", srv.Name)
	}
	return fmt.Sprintf("/%s/%s/", srv.Name, srv.Version)
}

func BuildRegisterPath(srv Server) string {
	return fmt.Sprintf("%s%s", BuildPrefix(srv), srv.Addr)
}

func SplitPath(path string) (Server, error) {
	srv := Server{}
	strs := strings.Split(path, "/")
	if len(strs) == 0 {
		return srv, errors.New("invalid path")
	}

	srv.Addr = strs[len(strs)-1]

	return srv, nil
}

func BuildResolverUrl(app string) string {
	return scheme + ":///" + app
}

func ParseVal(val []byte) (Server, error) {
	srv := Server{}
	err := json.Unmarshal(val, &srv)
	return srv, err
}

func Exist(l []resolver.Address, addr resolver.Address) bool {
	for key := range l {
		if l[key].Addr == addr.Addr {
			return true
		}
	}
	return false
}

func Remove(l []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for key := range l {
		if l[key].Addr == addr.Addr {
			l[key] = l[len(l)-1]
			return l[:len(l)-1], true
		}
	}
	return nil, false
}
