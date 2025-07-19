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
	config.InitConfig("job")

	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

func main() {
	//swagger gin api init
	runMode, exit := config.GetAppConfigValue[string]("common.server.runmode")
	if exit == false {
		log.Fatal("runmode not found")
	}
	gin.SetMode(*runMode)

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

	err = pb.RegisterProductsSearchServiceHandlerFromEndpoint(ctx, mux, "grpc.easy.dev:50052", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	err = pb.RegisterPriceSearchServiceHandlerFromEndpoint(ctx, mux, "grpc.easy.dev:50052", opts)
	if err != nil {
		log.Fatalf("failed to register gateway: %v", err)
	}

	g.GET("/swagger-doc/swagger.json", func(c *gin.Context) {
		c.File("./protos/services/search.swagger.json")
	})

	url := ginSwagger.URL("/swagger-doc/swagger.json")
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	g.Any("/v1/*any", gin.WrapH(mux))

	isTls, exit := config.GetAppConfigValue[bool]("common.server.tls")
	if exit == false {
		log.Fatalf("failed to get tls: %v", err)
	}
	addr, exit := config.GetAppConfigValue[string]("common.server.addr")
	if exit == false {
		log.Fatalf("failed to get addr: %v", err)
	}
	cert, exit := config.GetAppConfigValue[string]("common.server.cert")
	if exit == false {
		log.Fatalf("failed to get cert: %v", err)
	}
	key, exit := config.GetAppConfigValue[string]("common.server.key")
	if exit == false {
		log.Fatalf("failed to get key: %v", err)
	}

	if *isTls {
		log.Printf("开始监听服务器地址: %s\n", "https://"+*addr)
		err = g.RunTLS(*addr, *cert, *key)
		if err != nil {
			log.Fatalf("failed to run tls: %v", err)
		}
	} else {
		log.Printf("开始监听服务器地址: %s\n", "http://"+*addr)
		err = g.Run(*addr)
		if err != nil {
			log.Fatalf("failed to run tls: %v", err)
		}
	}

}
