package handler

import (
	"context"
	"user/internal/repository"
	"user/internal/service"
	"user/pkg/e"
)

type UserService struct {
	// 这里按要求嵌入 service.UnimplementedUserServiceServer
	*service.UnimplementedUserServiceServer
}

func NewUserService() *UserService {
	return &UserService{}
}

func (u UserService) UserLogin(ctx context.Context, req *service.UserRequest) (*service.UserDetailResponse, error) {
	var user repository.User
	resp := new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	err := user.ShowUserInfo(req)
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.UserDetail = repository.BuildUser(user)
	return resp, nil
}

func (u UserService) UserRegister(ctx context.Context, req *service.UserRequest) (*service.UserDetailResponse, error) {
	var user repository.User
	resp := new(service.UserDetailResponse)
	resp.Code = e.SUCCESS
	err := user.Create(req)
	if err != nil {
		resp.Code = e.ERROR
		return resp, err
	}
	resp.UserDetail = repository.BuildUser(user)
	return resp, nil
}

func (u UserService) UserLogout(ctx context.Context, req *service.UserRequest) (*service.UserDetailResponse, error) {
	resp := new(service.UserDetailResponse)
	return resp, nil
}
