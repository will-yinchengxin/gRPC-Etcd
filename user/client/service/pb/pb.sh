# https://developers.google.com/protocol-buffers/docs/gotutorial
# protoc -I . *.proto --go-grpc_out=.
# protoc:表示调用protoc进行代码生成
# -I          :表示对应搜索路径（import）
# .           :代表当前路径
# proto.proto :表示在路径下搜索的文件
# --go_out=   :表示生成代码格式（--go-grpc_out=.）
# plugins=    :表示对应插件这里是grpc，可以不写（protoc -I . proto.proto --go_out=:.）
# :.          :表示生成路径
#
# https://developers.google.com/protocol-buffers/docs/gotutorial

# syntax="proto3";
  #package pb;
  #option go_package = "/internal/service;service";
  #
  #message UserModel {
  #  // @inject_tag: json:"user_id"
  #  uint32 UserID=1;
  #  // @inject_tag: json:"user_name"
  #  string UserName=2;
  #  // @inject_tag: json:"nick_name"
  #  string NickName=3;
  #}
  #
  #//  option go_package = "path;name";
  #//  - path 表示生成的go文件的存放地址，会自动生成目录的。
  #//  - name 表示生成的go文件所属的包名
  #//
  #// 手动安装一下 protoc-go-inject-tag 库
  #// go get github.com/favadi/protoc-go-inject-tag
  #// 可以在proto文件中注入tag，然后在导出的时候相应的字段的tag就可以被修改掉, 否则使用默认 pb 的 tag
#

# 在 根目录下执行即可
protoc -I internal/service/pb    internal/service/pb/*.proto --go_out=.
protoc -I internal/service/pb    internal/service/pb/*.proto --go-grpc_out=.
