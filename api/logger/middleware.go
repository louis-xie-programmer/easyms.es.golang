package logger

import (
	"easyms-es/api/errno"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	"context"

	"google.golang.org/grpc/status"
)

var (
	validClientID     = "easy"
	validClientSecret = "ebsaekcHSan38yNVEKMJd6LfoMyv2KWG"
)

// 验证客户端授权码
// 参数:
//
//	clientID - 客户端ID
//	clientSecret - 客户端密钥
//
// 返回:
//
//	bool - 验证结果
func validateClient(clientID, clientSecret string) bool {
	return clientID == validClientID && clientSecret == validClientSecret
}

// gin 日志中间价件, 添加了客户端校验, 代码已迁移
//func GinLoggerMiddleware() gin.HandlerFunc {
//	return func(c *gin.Context) {
//		clientID := c.GetHeader("Client-ID")
//		clientSecret := c.GetHeader("Client-Secret")
//		userIP := c.GetHeader("User-Real-IP")
//		userAgent := c.GetHeader("User-Real-Agent")
//
//		startTime := time.Now()
//
//		if !validateClient(clientID, clientSecret) {
//			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
//			c.Error(fmt.Errorf("unauthorized"))
//		} else {
//			c.Next()
//		}
//
//		endTime := time.Now()
//		latency := int(endTime.Sub(startTime).Milliseconds())
//		statusCode := c.Writer.Status()
//		clientIP := c.ClientIP()
//		path := c.Request.URL.Path
//
//		var errStr string
//		if len(c.Errors) > 0 {
//			errStr = c.Errors.String()
//			zap.S().Error(errStr)
//		}
//
//		timestamp, _ := time.Parse("2006-01-02 15:04:05", endTime.Format("2006-01-02 15:04:05"))
//
//		logEntry := LogEntry{
//			Service:    "gin",
//			Method:     path,
//			ClientID:   clientID,
//			ClientIP:   clientIP,
//			UserIP:     userIP,
//			UserAgent:  userAgent,
//			StatusCode: statusCode,
//			Latency:    latency,
//			Timestamp:  timestamp,
//		}
//
//		LogAsync(logEntry)
//	}
//}

// GrpcLoggerUnaryInterceptor gRPC日志拦截器
// 实现客户端验证和日志记录功能
// 返回:
//
//	grpc.UnaryServerInterceptor - 一元拦截器
func GrpcLoggerUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var (
			resp interface{}
			err  error
		)

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			err = errors.New("missing metadata")
		}

		clientID := getMetadataValue(md, "Client-ID")
		clientSecret := getMetadataValue(md, "Client-Secret")

		userIP := getMetadataValue(md, "User-Real-IP")
		if userIP == "" {
			userIP = getMetadataValue(md, "X-Forwarded-For")
		}
		userAgent := getMetadataValue(md, "User-Real-Agent")
		if userAgent == "" {
			userAgent = getMetadataValue(md, "User-Agent")
		}

		clientIP := ""
		// 获取客户端IP地址
		p, ok := peer.FromContext(ctx)
		if ok {
			clientIP, _, _ = net.SplitHostPort(p.Addr.String())
		}

		if !validateClient(clientID, clientSecret) {
			err = errors.New("invalid client secret")
		}

		startTime := time.Now()

		if err == nil {
			resp, err = handler(ctx, req)
		}

		if err != nil {
			err = errno.HandleError(err)
		}

		endTime := time.Now()

		latency := int(endTime.Sub(startTime).Milliseconds())
		params, _ := json.Marshal(req)
		url := info.FullMethod
		st, _ := status.FromError(err)

		timestamp, _ := time.Parse("2006-01-02 15:04:05", endTime.Format("2006-01-02 15:04:05"))
		logEntry := LogEntry{
			Service:    "grpc",
			Method:     url,
			ClientID:   clientID,
			ClientIP:   clientIP,
			UserIP:     userIP,
			UserAgent:  userAgent,
			StatusCode: int(st.Code()),
			Latency:    latency,
			Timestamp:  timestamp,
			Error:      st.Message(),
			Params:     string(params),
		}

		LogAsync(logEntry)

		return resp, err
	}
}

// getMetadataValue 获取metadata值
// 参数:
//
//	md - metadata对象
//	key - 键值
//
// 返回:
//
//	string - 值
func getMetadataValue(md metadata.MD, key string) string {
	if values := md.Get(strings.ToLower(key)); len(values) > 0 {
		return values[0]
	}
	return ""
}
