package protocol

import "fmt"

// 常用错误类型
var (
	ErrNotInteger  = &Error{Message: "ERR value is not an integer"}
	ErrNotFloat    = &Error{Message: "ERR value is not a float"}
	ErrSyntax      = &Error{Message: "ERR syntax error"}
	ErrWrongNumber = &Error{Message: "ERR wrong number of arguments"}
	ErrUnknown     = &Error{Message: "ERR unknown command"}
)

// MakeError 创建错误
func MakeError(format string, args ...interface{}) *Error {
	return &Error{Message: fmt.Sprintf(format, args...)}
}
