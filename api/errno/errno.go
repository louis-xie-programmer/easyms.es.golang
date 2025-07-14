package errno

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

// 预定义错误类型
var (
	ParamError    = "param error"    // 参数错误标识
	NotfoundError = "not found error"  // 资源未找到错误标识
	TimeoutError  = "timeout error"    // 操作超时错误标识
	ESError       = "es error"         // Elasticsearch相关错误标识
)

// MapErrorToGRPCCode 将错误信息映射到gRPC状态码
// 参数:
//   err - 错误对象
// 返回:
//   codes.Code - gRPC状态码
func MapErrorToGRPCCode(err error) codes.Code {
	switch {
	case err.Error() == "invalid client secret":
		return 403
	case err.Error() == "missing metadata":
		return 403
	case strings.HasPrefix(err.Error(), NotfoundError):
		return 404
	case strings.HasPrefix(err.Error(), ParamError):
		return 2001
	case strings.HasPrefix(err.Error(), TimeoutError):
		return 2002
	case strings.HasPrefix(err.Error(), ESError):
		return 2003
	default:
		return 500
	}
}

// HandleError 统一处理错误并返回gRPC格式错误
// 参数:
//   err - 错误对象
// 返回:
//   error - gRPC格式错误
func HandleError(err error) error {
	if err == nil {
		return nil
	}

	code := MapErrorToGRPCCode(err)
	return status.Error(code, err.Error())
}
