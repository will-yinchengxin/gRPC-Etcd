## 初试 proto
目录结构
````
├── client
│   └── client.go
│   └── stream
│       └── client.go
├── go.mod
├── go.sum
├── proto
│   ├── userServer.pb.go
│   └── userServer.proto
└── server
│   └── server.go
│   └── stream
│       └── server.go
````
### 普通调用
````
$ go run client/client.go
23
id:1  name:"test"  age:31

$ go run server/server.go
userServer grpc services start success
age:31name:"test"
age:31name:"test"
age:31name:"test"
age:31name:"test"
````
### Client-side streaming RPC 客户端/服务端 流模式调用
````
$ go run server/stream/server.go
article stream Server grpc services start success


$ go run client/stream/client.go
stream.rev aid: 2, author: jack, title: title_go_2, context: content_go_2
.....
````

### 双向流模式调用
双向模式顾名思义，就是client和server都是流水模式，2边一起流水。代码均在各自的 `DeleteArticle` 方法中

### TLS加密通讯
https的核心逻辑：
````
server 采用非对称加密，生成一个公钥public1和私钥private1
server 把公钥public1传给client
client 采用对称加密生成1个秘钥A （或者2个秘钥A，内容都是一样）
client 用server给自己的公钥public1加密自己生成的对称秘钥A。生成了一个秘钥B.
client 把秘钥B传给server。
client 用秘钥A加密需要传输的数据Data，并传给server。
server 收到秘钥B后，用自己的私钥private1解开了，得到了秘钥A。
server 收到加密后的data后，用秘钥A解开了，获得了元素数据。
````
简而言之，就是采用非对称加密+对称加密的方式。其中，对称加密产生的秘钥，是既可以加密，又可以解密的，加密解密速度很快。而采用非对称加密，则不可以，
必须公钥解密私钥，或者私钥加密公钥，加解密速度慢。这样一个组合，就可以保障数据得到加密，又不会影响速度。

首先，我们要生成server的公钥public1和私钥private1。那就得用到openssl命令了。需要注意的是go在1.15版本，X509无法使用了，需要用Sans算法代替。
````
[root@ha1 openssl]# openssl genrsa -out proto.key 2048
Generating RSA private key, 2048 bit long modulus
.+++
.....+++
e is 65537 (0x10001)

[root@ha1 openssl]# ls
proto.key

[root@ha1 openssl]# openssl req -new -x509 -key proto.key -out proto.pem -days 365
You are about to be asked to enter information that will be incorporated
into your certificate request.
What you are about to enter is what is called a Distinguished Name or a DN.
There are quite a few fields but you can leave some blank
For some fields there will be a default value,
If you enter '.', the field will be left blank.
-----
Country Name (2 letter code) [XX]:CN
State or Province Name (full name) []:GD
Locality Name (eg, city) [Default City]:SZ
Organization Name (eg, company) [Default Company Ltd]:IT
Organizational Unit Name (eg, section) []:it
Common Name (eg, your name or your server's hostname) []:prototest
Email Address []:prototest@qq.com

[root@ha1 openssl]# ls
proto.key  proto.pem
````

