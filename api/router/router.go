package router

import (
	pb "easyms-es/protos/services"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// InitGrpc 初始化并注册gRPC服务
// 参数:
//
//	grpcServer - gRPC服务器实例
func InitGrpc(grpcServer *grpc.Server) {
	// 注册产品搜索服务
	pb.RegisterProductsSearchServiceServer(grpcServer, &ProductEsServer{})
	pb.RegisterPriceSearchServiceServer(grpcServer, &PriceEsServer{})
	// 启用反射服务（用于调试）
	reflection.Register(grpcServer)
}
