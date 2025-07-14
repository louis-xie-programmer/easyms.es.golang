package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"easyms-es/api/grpcquic"
	ms "easyms-es/protos/messages"
	pb "easyms-es/protos/services"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
)

func EasyGrpcQUICClient() error {
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
		NextProtos:   []string{"h3"},
	}
	creeds := grpcquic.NewCredentials(tlsConfig)

	// Connect to gRPC Service Server
	dialer := grpcquic.NewQuickDialer(tlsConfig)

	grpcOpts := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(creeds),
	}

	conn, err := grpc.Dial("grpc.easy.dev:9443", grpcOpts...)
	if err != nil {
		log.Printf("QuicClient: failed to grpc.Dial. %s", err.Error())
		return err
	}
	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Printf("QuicClient: failed to close - grpc.Dial. %s", err.Error())
		}
	}(conn)

	c := pb.NewProductsSearchServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Analyze(ctx, &ms.ProductSearchParam{KeyWord: "iphone 14"})
	if err != nil {
		log.Printf("QuicClient: could not greet. %v", err)
		return err
	}
	log.Printf("QuicClient: Greeting=%v", r.Tokens)

	return nil
}

func main() {
	if err := EasyGrpcQUICClient(); err != nil {
		log.Printf("failed to QUIC Client. %s", err.Error())
		return
	}
	select {}
}
