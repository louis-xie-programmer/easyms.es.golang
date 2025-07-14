// 基于quic的证书验证模块
package grpcquic

import (
	"context"
	"crypto/tls"
	"net"

	"google.golang.org/grpc/credentials"
)

// Info 传输信息封装
// 包含底层QUIC连接
type Info struct {
	conn *Conn
}

// NewInfo 创建传输信息实例
// 参数:
//
//	c - QUIC连接实例
//
// 返回:
//
//	*Info - 传输信息对象
func NewInfo(c *Conn) *Info {
	return &Info{c}
}

// AuthType 获取认证类型
// 返回:
//
//	string - 认证类型标识
func (i *Info) AuthType() string {
	return "quic-tls"
}

// Conn 获取网络连接
// 返回:
//
//	net.Conn - 网络连接实例
func (i *Info) Conn() net.Conn {
	return i.conn
}

// Credentials 传输凭证封装
// 实现gRPC TransportCredentials接口
type Credentials struct {
	config           *tls.Config
	isQUICConnection bool
	serverName       string

	cred credentials.TransportCredentials
}

// NewCredentials 创建凭证实例
// 参数:
//
//	config - TLS配置
//
// 返回:
//
//	credentials.TransportCredentials - 凭证接口实例
func NewCredentials(config *tls.Config) credentials.TransportCredentials {
	cred := credentials.NewTLS(config)
	return &Credentials{
		cred:   cred,
		config: config,
	}
}

// ClientHandshake 客户端握手
// 参数:
//
//	ctx - 上下文
//	authority - 权限标识
//	conn - 网络连接
//
// 返回:
//
//	net.Conn - 加密连接
//	credentials.AuthInfo - 认证信息
//	error - 错误信息
func (pt *Credentials) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if c, ok := conn.(*Conn); ok {
		pt.isQUICConnection = true
		return conn, NewInfo(c), nil
	}

	return pt.cred.ClientHandshake(ctx, authority, conn)
}

// ServerHandshake 服务端握手
// 参数:
//
//	conn - 网络连接
//
// 返回:
//
//	net.Conn - 加密连接
//	credentials.AuthInfo - 认证信息
//	error - 错误信息
func (pt *Credentials) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if c, ok := conn.(*Conn); ok {
		pt.isQUICConnection = true
		aInfo := NewInfo(c)
		return conn, aInfo, nil
	}

	return pt.cred.ServerHandshake(conn)
}

// Info 获取协议信息
// 返回:
//
//	credentials.ProtocolInfo - 协议信息
func (pt *Credentials) Info() credentials.ProtocolInfo {
	if pt.isQUICConnection {
		return credentials.ProtocolInfo{
			ProtocolVersion:  "/quic/1.0.0",
			SecurityProtocol: "quic-tls",
			ServerName:       pt.serverName,
		}
	}

	return pt.cred.Info()
}

// Clone 创建凭证副本
// 返回:
//
//	credentials.TransportCredentials - 新凭证实例
func (pt *Credentials) Clone() credentials.TransportCredentials {
	return &Credentials{
		config: pt.config.Clone(),
		cred:   pt.cred.Clone(),
	}
}

// OverrideServerName 覆盖服务器名称
// 参数:
//
//	name - 新服务器名称
//
// 返回:
//
//	error - 错误信息
func (pt *Credentials) OverrideServerName(name string) error {
	pt.serverName = name
	return pt.cred.OverrideServerName(name)
}
