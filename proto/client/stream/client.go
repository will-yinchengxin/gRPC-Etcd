package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io"
	"log"
	"proto/proto"
)

var client proto.ArticleServerClient

func main() {

	connect, err := grpc.Dial("127.0.0.1:9527", grpc.WithInsecure())

	if err != nil {
		log.Fatal("connect grpc fail")
	}

	defer connect.Close()

	client = proto.NewArticleServerClient(connect)

	SaveArticle()
	//GetArticleInfo()
	//DeleteArticle()

}

func SaveArticle() {
	//定义一组数据
	SaveList := map[string]proto.ArticleParam{
		"1": {Author: &proto.Author{Author: "tony"}, Title: &proto.Title{Title: "title1"}, Content: &proto.Content{Content: "content1"}},
		"2": {Author: &proto.Author{Author: "jack"}, Title: &proto.Title{Title: "title2"}, Content: &proto.Content{Content: "content2"}},
		"3": {Author: &proto.Author{Author: "tom"}, Title: &proto.Title{Title: "title3"}, Content: &proto.Content{Content: "content3"}},
		"4": {Author: &proto.Author{Author: "boby"}, Title: &proto.Title{Title: "title4"}, Content: &proto.Content{Content: "content4"}},
	}

	//先调用函数
	stream, err := client.SaveArticle(context.Background())

	if err != nil {
		log.Fatal("SaveArticle grpc fail", err.Error())
	}

	//再循环发送
	for _, info := range SaveList {
		err = stream.Send(&info)
		if err != nil {
			log.Fatal("SaveArticle Send info fail", err.Error())
		}
	}

	//发送关闭新号，并且获取返回值
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("SaveArticle CloseAndRecv fail", err.Error())
	}

	fmt.Printf("resp: id = %d", resp.GetId())
}

func SaveArticle2() {
	//定义一组数据
	SaveInfo := proto.ArticleParam{
		Author: &proto.Author{Author: "mark"}, Title: &proto.Title{Title: "title5"}, Content: &proto.Content{Content: "content5"},
	}

	//先调用函数
	stream, err := client.SaveArticle(context.Background())

	if err != nil {
		log.Fatal("SaveArticle grpc fail", err.Error())
	}

	//发送
	err = stream.Send(&SaveInfo)
	if err != nil {
		log.Fatal("SaveArticle Send info fail", err.Error())
	}

	////发送关闭新号，并且获取返回值
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("SaveArticle CloseAndRecv fail", err.Error())
	}

	fmt.Printf("resp: id = %d", resp.GetId())
}

func GetArticleInfo() {
	Aid := proto.Aid{
		Id: 2,
	}

	//请求
	stream, err := client.GetArticleInfo(context.Background(), &Aid)

	if err != nil {
		log.Fatal("GetArticleInfo grpc fail", err.Error())
	}

	//循环接受server流发来数据
	for {
		r, err := stream.Recv()

		if err == io.EOF {
			fmt.Println("读取数据结束")
			break
		}

		if err != nil {
			log.Fatal("GetArticleInfo Recv fail", err.Error())
		}

		fmt.Printf("stream.rev aid: %d, author: %s, title: %s, context: %s\n", r.GetId(), r.GetAuthor(), r.GetTitle(), r.GetContent())

	}
}

// 双向流
func DeleteArticle() {
	//链接rpc
	stream, err := client.DeleteArticle(context.Background())

	if err != nil {
		log.Fatal("DeleteArticle grpc fail", err.Error())
	}

	for i := 0; i < 6; i++ {

		//先发
		err = stream.Send(&proto.Aid{Id: int32(i)})
		if err != nil {
			log.Fatal("DeleteArticle Send fail", err.Error())
		}

		//再收
		r, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal("GetArticleInfo Recv fail", err.Error())
		}

		fmt.Printf("stream.rev status: %v\n", r.GetCode())
	}

	//发送结束
	_ = stream.CloseSend()
}
