package handler

import (
	"api-gateway/pkg/util"
	"errors"
)

func PanicIfUserError(err error) {
	if err != nil {
		err = errors.New("userService--" + err.Error())
		util.LogrusObj.Info(err)
		panic(err)
	}
}
