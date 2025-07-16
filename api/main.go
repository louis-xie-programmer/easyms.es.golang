package main

import (
	"easyms-es/api/logger"
	"easyms-es/api/router"
	"easyms-es/model"
	"easyms-es/service/prices"
	"easyms-es/service/products"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"log"
	"net"
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

	//http3 + grpc init 注意尚未解决net6客户端兼容问题,
	go func() {
		err := router.EasyGrpcQUICServer("grpc.easy.bom:50051", "./certs/server.crt", "./certs/server.key")
		if err != nil {
			log.Printf("failed to Echo QUIC Server. %s", err.Error())
			return
		}
		log.Printf("gRPC server listening at %s", "grpc.easy.bom:50051")
	}()

	cert, err := credentials.NewServerTLSFromFile(certFilePath, keyFilePath)
	if err != nil {
		log.Fatalf("failed to load TLS certificates: %v", err)
	}

	// keepalive 设置
	var keepAliveArgs = keepalive.ServerParameters{
		Time:              10 * time.Second,
		Timeout:           30 * time.Second,
		MaxConnectionIdle: 3 * time.Minute,
	}

	//grpc init
	grpcServer := grpc.NewServer(
		grpc.Creds(cert),
		grpc.KeepaliveParams(keepAliveArgs),
		grpc.UnaryInterceptor(logger.GrpcLoggerUnaryInterceptor()),
	)
	router.InitGrpc(grpcServer)
	listen, err := net.Listen("tcp", "grpc.easy.bom:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening at %v", listen.Addr())

	if err := grpcServer.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	logger.CloseLogging()
}
