package repository

import (
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"user/internal/service"
)

type User struct {
	UserId         int64  `json:"user_id" gorm:"user_id"`
	UserName       string `json:"user_name" gorm:"user_name"`
	NickName       string `json:"nick_name" gorm:"nick_name"`
	PasswordDigest string `json:"password_digest" gorm:"password_digest"`
}

// TableName 表名称
func (*User) TableName() string {
	return "user"
}

const (
	PassWordCost = 12 // 密码加密难度
)

// 加密密码
func (u *User) SetPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), PassWordCost)
	if err != nil {
		return err
	}
	u.PasswordDigest = string(bytes)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordDigest), []byte(password))
	return err == nil
}

func (u *User) CheckUserExist(req *service.UserRequest) bool {
	if err := DB.Where("user_name = ?", req.UserName).
		First(&u).Error; err == gorm.ErrRecordNotFound {
		return false
	}
	return true
}

func (u *User) ShowUserInfo(req *service.UserRequest) (err error) {
	if exist := u.CheckUserExist(req); exist {
		return nil
	}
	return errors.New("UserName Not Exist")
}

func (*User) Create(req *service.UserRequest) error {
	var user User
	var count int64
	DB.Where("user_name=?", req.UserName).Count(&count)
	if count != 0 {
		return errors.New("UserName Exist")
	}
	user = User{
		UserName: req.UserName,
		NickName: req.NickName,
	}
	_ = user.SetPassword(req.Password)
	if err := DB.Create(&user).Error; err != nil {
		fmt.Println("err ====> ", err)
		return err
	}
	return nil
}

func BuildUser(item User) *service.UserModel {
	return &service.UserModel{
		UserName: item.UserName,
		UserID:   uint32(item.UserId),
		NickName: item.NickName,
	}
}
