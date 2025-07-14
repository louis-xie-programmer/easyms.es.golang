package main

import (
	"crypto/tls"
	"easyms-es/api/logger"
	"easyms-es/api/router"
	"easyms-es/model"
	"easyms-es/service/prices"
	"easyms-es/service/products"
	"github.com/quic-go/quic-go/http3"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"runtime"
	"time"
)

import (
	"easyms-es/config"
	"easyms-es/easyes"
)

var (
	// TLS证书路径、日志数据库连接等配置变量
	certFilePath          = ""
	keyFilePath           = ""
	logConnString         = ""
	esProductIndexName    = ""
	esPriceStockIndexName = ""
	timeout               = 30
)

// init 初始化应用配置和核心组件
// 包含配置加载、证书设置、日志系统初始化和Elasticsearch存储服务创建
func init() {
	// 初始化配置系统
	config.InitConfig()

	// 加载主配置文件
	configName := "config"
	config.CreateConfig(configName, "./conf/api/"+configName+".yaml")

	// 读取服务器证书路径和日志数据库连接字符串
	certFilePath = config.GetSyncConfig(configName, "common.server.cert")
	keyFilePath = config.GetSyncConfig(configName, "common.server.key")
	logConnString = config.GetSyncConfig(configName, "common.logdb.connstring")
	
	// 读取Elasticsearch索引名称和超时设置
	esProductIndexName = config.GetSyncConfig(configName, "common.elasticsearch.productindex")
	esPriceStockIndexName = config.GetSyncConfig(configName, "common.elasticsearch.pricestockindex")
	timeout = config.GetSyncConfig_Type[int](configName, "common.elasticsearch.timeout")

	// 使用数据库连接初始化日志系统
	err := logger.InitLogger(logConnString)

	if err != nil {
		log.Fatalf(err.Error())
	}

	// 初始化模型层静态变量
	model.EsProductIndexName = esProductIndexName
	model.EsPriceIndexName = esPriceStockIndexName

	// 初始化Elasticsearch存储服务
	store1, err := easyes.NewStore(easyes.StoreConfig{
		IndexName: model.EsProductIndexName,
		Timeout:   time.Duration(timeout) * time.Second,
	})
	if err != nil {
		log.Fatalf(err.Error())
	}
	products.ProductStore = *store1
	store2, err := easyes.NewStore(easyes.StoreConfig{
		IndexName: model.EsPriceIndexName,
		Timeout:   time.Duration(timeout) * time.Second,
	})
	if err != nil {
		log.Fatalf(err.Error())
	}
	prices.PriceStore = *store2
}

// main 函数启动微服务
// 启动HTTP/2和HTTP/3服务并注册gRPC接口
func main() {
	// 启动pprof性能监控
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Printf(http.ListenAndServe(":6060", nil).Error())
	}()

	// 创建HTTP路由复用器
	mux := http.NewServeMux()

	// 创建带拦截器的gRPC服务器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger.GrpcLoggerUnaryInterceptor()),
	)

	// 注册gRPC服务到路由
	router.InitGrpc(grpcServer)
	mux.Handle("/", grpcServer)

	// 配置HTTPS服务器参数
	server := &http.Server{
		Addr:    ":50051",
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	// 启动HTTP/2服务
	go func() {
		log.Printf("Starting HTTP/2 server on :50051")
		log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
	}()

	// 启动HTTP/3服务
	log.Printf("Starting HTTP/3 server on :50051")
	log.Fatal(http3.ListenAndServeQUIC(":50051", certFilePath, keyFilePath, mux))
}
