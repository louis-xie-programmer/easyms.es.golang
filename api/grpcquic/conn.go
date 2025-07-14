// Package grpcquic 提供gRPC与HTTP/3协议的封装
// 实现基于QUIC协议的网络连接管理
package grpcquic

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/quic-go/quic-go"
)

// Conn QUIC连接封装
// 包含底层QUIC连接和流式传输通道
type Conn struct {
	conn   quic.Connection
	stream quic.Stream
}

// NewConn 创建新的QUIC连接
// 参数:
//   conn - QUIC连接实例
// 返回:
//   net.Conn - 封装后的连接
//   error - 错误信息
func NewConn(conn quic.Connection) (net.Conn, error) {
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{conn, stream}, nil
}

// Read 从流中读取数据
// 参数:
//   b - 数据缓冲区
// 返回:
//   int - 读取字节数
//   error - 错误信息
func (c *Conn) Read(b []byte) (n int, err error) {
	return c.stream.Read(b)
}

// Write 将数据写入流
// 参数:
//   b - 待写入数据
// 返回:
//   int - 写入字节数
//   error - 错误信息
func (c *Conn) Write(b []byte) (n int, err error) {
	return c.stream.Write(b)
}

// Close 关闭连接
// 返回:
//   error - 错误信息
func (c *Conn) Close() error {
	err := c.stream.Close()
	if err != nil {
		return err
	}
	return c.conn.CloseWithError(0, "")
}

// LocalAddr 获取本地地址
// 返回:
//   net.Addr - 本地地址信息
func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr 获取远程地址
// 返回:
//   net.Addr - 远程地址信息
func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// SetDeadline 设置连接截止时间
// 参数:
//   t - 截止时间
// 返回:
//   error - 错误信息
func (c *Conn) SetDeadline(t time.Time) error {
	return c.stream.SetDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.stream.SetReadDeadline(t)

}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.stream.SetWriteDeadline(t)
}

type Listener struct {
	ql quic.Listener
}

// Listen 创建监听器实例
// 参数:
//   ql - QUIC监听器
// 返回:
//   net.Listener - 封装后的监听器
func Listen(ql quic.Listener) net.Listener {
	return &Listener{ql}
}

// Accept 接受新连接
// 返回:
//   net.Conn - 新连接实例
//   error - 错误信息
func (l *Listener) Accept() (net.Conn, error) {
	conn, err := l.ql.Accept(context.Background())
	if err != nil {
		return nil, err
	}

	stream, err := conn.AcceptStream(context.Background())
	if err != nil {
		return nil, err
	}

	return &Conn{conn, stream}, nil
}

// Close 关闭监听器
// 返回:
//   error - 错误信息
func (l *Listener) Close() error {
	return l.ql.Close()
}

// Addr 获取监听地址
// 返回:
//   net.Addr - 监听地址
func (l *Listener) Addr() net.Addr {
	return l.ql.Addr()
}

// NewQuickDialer 创建快速拨号器
// 参数:
//   conf - TLS配置
// 返回:
//   func - 拨号函数
func NewQuickDialer(conf *tls.Config) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, target string) (net.Conn, error) {
		conn, err := quic.DialAddr(ctx, target, conf, &quic.Config{})
		if err != nil {
			return nil, err
		}

		return NewConn(conn)
	}
}
