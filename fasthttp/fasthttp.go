package fasthttp

import (
	"bytes"
	"crypto/tls"
	"io"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

// Transport 自定义传输层实现
// 提供TLS配置和高性能fasthttp客户端
type Transport struct {
	TLSClientConfig *tls.Config // TLS配置
	Client          *fasthttp.Client // fasthttp客户端实例
}

// NewTransport 创建带TLS配置的传输层
// 参数:
//   tlsConfig - TLS配置
// 返回:
//   *Transport - 自定义传输层实例
func NewTransport(tlsConfig *tls.Config) *Transport {
	return &Transport{
		TLSClientConfig: tlsConfig,
		Client: &fasthttp.Client{
			TLSConfig:                     tlsConfig,
			ReadTimeout:                   15 * time.Second,
			WriteTimeout:                  15 * time.Second,
			MaxIdleConnDuration:           30 * time.Second,
			NoDefaultUserAgentHeader:      true,
			DisableHeaderNamesNormalizing: true,
			DisablePathNormalizing:        true,
			Dial: (&fasthttp.TCPDialer{
				Concurrency:      4096,
				DNSCacheDuration: time.Hour,
			}).Dial,
		},
	}
}

// RoundTrip 执行HTTP请求
// 参数:
//   req - HTTP请求
// 返回:
//   *http.Response - HTTP响应
//   error - 错误信息
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	freq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(freq)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := t.copyRequest(freq, req); err != nil {
		return nil, err
	}

	if err := t.Client.Do(freq, resp); err != nil {
		return nil, err
	}

	res := &http.Response{Header: make(http.Header)}
	t.copyResponse(res, resp)

	return res, nil
}

// copyRequest 将标准库请求转换为fasthttp请求
// 参数:
//   dst - fasthttp请求对象
//   src - 标准库请求对象
// 返回:
//   error - 错误信息
func (t *Transport) copyRequest(dst *fasthttp.Request, src *http.Request) error {
	dst.SetHost(src.Host)
	dst.Header.SetMethod(src.Method)

	if src.URL.RawPath != "" {
		dst.SetRequestURI(src.URL.RawPath)
	} else {
		dst.SetRequestURI(src.URL.String())
	}

	for k, vv := range src.Header {
		for _, v := range vv {
			dst.Header.Add(k, v)
		}
	}

	if src.Body != nil {
		body, err := io.ReadAll(src.Body)
		if err != nil {
			return err
		}
		dst.SetBody(body)
		_ = src.Body.Close()
		src.Body = io.NopCloser(bytes.NewReader(body))
	}

	return nil
}

// copyResponse 将fasthttp响应转换为标准库响应
// 参数:
//   dst - 标准库响应对象
//   src - fasthttp响应对象
func (t *Transport) copyResponse(dst *http.Response, src *fasthttp.Response) {
	dst.StatusCode = src.StatusCode()

	src.Header.VisitAll(func(k, v []byte) {
		dst.Header.Set(string(k), string(v))
	})

	dst.Body = io.NopCloser(bytes.NewReader(src.Body()))
}
