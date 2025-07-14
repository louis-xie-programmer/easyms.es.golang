package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"easyms-es/config"
	pb "easyms-es/protos/services"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {}
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.9.0",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

var (
	validClientID     = "easy"
	validClientSecret = "obsaekcHSan38yNVEKMJd6LfoMyv2KWG"
)

func init() {
	//初始化配置文件数组
	config.InitConfig()

	//追加主配置文件
	configName := "config"
	config.CreateConfig(configName, "./conf/api/"+configName+".yaml")

	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

func main() {
	//swagger gin api init
	gin.SetMode(config.GetSyncConfig("", "common.server.runmode"))

	g := gin.Default()

	mux := runtime.NewServeMux()
	// 加载 gRPC 服务的 TLS 凭证
	// 加载客户端证书
	certificate, err := tls.LoadX509KeyPair("./certs/client.crt", "./certs/client.key")
	if err != nil {
		log.Fatalf("failed to load client certificate: %v", err)
	}

	// 加载 CA 证书
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("./certs/ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA certificate: %v", err)
	}

	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatalf("failed to append CA certificate")
	}

	// 配置 TLS 认证
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      certPool,
	}

	cred := credentials.NewTLS(tlsConfig)

	unaryInterceptor := func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// 创建上下文，并设置元数据（Client-ID 和 Client-Secret）
		md := metadata.New(map[string]string{
			"Client-ID":     validClientID,
			"Client-Secret": validClientSecret,
		})
		// 将 metadata 附加到 context
		ctx = metadata.NewOutgoingContext(ctx, md)
		// 继续调用 gRPC 方法
		return invoker(ctx, method, req, reply, cc, opts...)
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(cred),
		grpc.WithUnaryInterceptor(unaryInterceptor)}

	// 创建上下文，并设置元数据（Client-ID 和 Client-Secret）
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err = pb.RegisterProductsSearchServiceHandlerFromEndpoint(ctx, mux, "grpc.easy.dev:50051", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	err = pb.RegisterPriceSearchServiceHandlerFromEndpoint(ctx, mux, "grpc.easy.dev:50051", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	g.GET("/swagger-doc/swagger.json", func(c *gin.Context) {
		c.File("./protos/services/search.swagger.json")
	})

	url := ginSwagger.URL("/swagger-doc/swagger.json")
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	g.Any("/v1/*any", gin.WrapH(mux))

	isTls := config.GetSyncConfig_Type[bool]("", "common.server.tls")

	if isTls {
		log.Printf("开始监听服务器地址: %s\n", "https://"+config.GetSyncConfig("", "common.server.addr"))
		err = g.RunTLS(config.GetSyncConfig("", "common.server.addr"), config.GetSyncConfig("", "common.server.cert"), config.GetSyncConfig("", "common.server.key"))
		if err != nil {
			log.Fatalf("failed to run tls: %v", err)
		}
	} else {
		log.Printf("开始监听服务器地址: %s\n", "http://"+config.GetSyncConfig("", "common.server.addr"))
		err = g.Run(config.GetSyncConfig("", "common.server.addr"))
		if err != nil {
			log.Fatalf("failed to run tls: %v", err)
		}
	}

}
