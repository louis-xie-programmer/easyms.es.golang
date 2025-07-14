package router

import (
	"crypto/tls"
	"crypto/x509"
	qnet "easyms-es/api/grpcquic"
	pb "easyms-es/protos/services"
	"github.com/quic-go/quic-go"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
)

// EasyGrpcQUICServer http3 服务初始化，
func EasyGrpcQUICServer(addr, certFile, keyFile string) error {
	log.Println("starting echo QUICServer")

	// 加载服务端证书和密钥
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("failed to load server certificate: %v", err)
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

	// 创建 TLS 配置
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		NextProtos:   []string{"h3"},
	}

	ql, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		log.Printf("QUICServer: failed to ListenAddr. %s", err.Error())
		return err
	}
	defer ql.Close()

	listener := qnet.Listen(*ql)

	s := grpc.NewServer()
	pb.RegisterProductsSearchServiceServer(s, &ProductEsServer{})
	pb.RegisterPriceSearchServiceServer(s, &PriceEsServer{})
	//reflection.Register(s)
	log.Printf("QUICServer: listening at %v", listener.Addr())

	if err := s.Serve(listener); err != nil {
		log.Printf("QUICServer: failed to serve. %v", err)
		return err
	}

	log.Println("stopping echo QUICServer")
	return nil
}
