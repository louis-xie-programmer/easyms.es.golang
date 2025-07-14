package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	ms "easyms-es/protos/messages"
	pb "easyms-es/protos/services"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

var (
	validClientID     = "easy"
	validClientSecret = "obsaekcHSan38yNVEKMJd6LfoMyv2KWG"
)

func EasyGrpcClient() error {
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

	conn, err := grpc.NewClient("grpc.easy.dev:50051", grpc.WithTransportCredentials(cred))
	if err != nil {
		log.Printf("failed to connect to gRPC server: %v", err)
		return err
	}
	defer func() {
		if err = conn.Close(); err != nil {
			log.Printf("failed to close gRPC connection: %v", err)
		}
	}()

	client := pb.NewProductsSearchServiceClient(conn)

	// 创建上下文，并设置元数据（Client-ID 和 Client-Secret）
	md := metadata.New(map[string]string{
		"Client-ID":     validClientID,
		"Client-Secret": validClientSecret,
	})
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 设置超时
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	resp, err := client.Analyze(ctx, &ms.ProductSearchParam{KeyWord: "iphone 14 pro"})
	if err != nil {
		log.Fatalf("failed to call gRPC method: %v", err)
	}

	fmt.Printf("Response from server: %v\n", resp)

	return nil
}

func main() {
	if err := EasyGrpcClient(); err != nil {
		log.Printf("failed to QUIC Client. %s", err.Error())
		return
	}
	select {}
}
