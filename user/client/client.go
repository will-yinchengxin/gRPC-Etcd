package main

import (
	"client/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

func main() {
	r := gin.Default()

	rr := r.Group("/rpc")
	rr.GET("/userLogin", func(c *gin.Context) {
		sayHello(c)
	})

	//rr.GET("/UserRegister", func(c *gin.Context) {
	//	sayHello(c)
	//})
	//
	//rr.GET("/UserLogout", func(c *gin.Context) {
	//	sayHello(c)
	//})

	// Run http server
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}

func sayHello(c *gin.Context) {
	// Set up a connection to the server.
	conn, err := grpc.Dial("127.0.0.1:10001", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := service.NewUserServiceClient(conn)
	req := &service.UserRequest{UserName: "will", NickName: "will", Password: "123456", PasswordConfirm: "123456"}
	res, err := client.UserRegister(c, req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result": fmt.Sprint(res.GetUserDetail()),
		"code":   fmt.Sprint(res.Code),
	})

}
