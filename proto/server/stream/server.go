package main

//流式服务
import (
  "fmt"
  "proto/proto"
  "google.golang.org/grpc"
  "log"
  "net"
  "math/rand"
  "io"
)

type ArticleServerServer interface {
  SaveArticle(proto.ArticleServer_SaveArticleServer) error
  GetArticleInfo(*proto.Aid, proto.ArticleServer_GetArticleInfoServer) error
  DeleteArticle(proto.ArticleServer_DeleteArticleServer) error
}

type StreamArticleServer struct {
}

func (server *StreamArticleServer) SaveArticle(stream proto.ArticleServer_SaveArticleServer) error {
  for {
      id := rand.Int31n(100)
      r, err := stream.Recv()
      if err == io.EOF {
        fmt.Println("读取数据结束")
        res := &proto.Aid{Id: id}
        return stream.SendAndClose(res)
      }

      if err != nil {
        return err
      }

      fmt.Printf("stream.rev author: %s, title: %s, context: %s", r.Author.GetAuthor(), r.Title.GetTitle(), r.Content.GetContent())
    }
}

func (server *StreamArticleServer) GetArticleInfo(aid *proto.Aid, stream proto.ArticleServer_GetArticleInfoServer) error {

  for i := 0; i < 6; i++ {
    id := strconv.Itoa(int(aid.GetId()))
    err := stream.Send(&proto.ArticleInfo{
      Id:      aid.GetId(),
      Author:  "jack",
      Title:   "title_go_" + id,
      Content: "content_go_" + id,
    })

    if err != nil {
      return err
    }
  }
  fmt.Println("发送完毕")
  return nil
}

// 双向流
func (server *StreamArticleServer) DeleteArticle(stream proto.ArticleServer_DeleteArticleServer) error {
  for {

      //循环接收client发送的流数据
      r, err := stream.Recv()
      if err == io.EOF {
        fmt.Println("read done!")
        return nil
      }

      if err != nil {
        return err
      }

      fmt.Printf("stream.rev aid: %d\n", r.GetId())

      //循环发流数据给client
      err = stream.Send(&proto.Status{Code: true})

      if err != nil {
        return err
      }

      //fmt.Println("send done!")
    }
}

func main() {

  listen, err := net.Listen("tcp", "127.0.0.1:9527")
  if err != nil {
    log.Fatalf("tcp listen failed:%v", err)
  }

  server := grpc.NewServer()

  proto.RegisterArticleServerServer(server, &StreamArticleServer{})
  fmt.Println("article stream Server grpc services start success")

  _ = server.Serve(listen)

}