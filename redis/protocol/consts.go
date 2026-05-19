package protocol

// RESP协议常量
const (
	// 换行符
	CRLF = "\r\n"

	// 类型标识
	SIMPLE_STRING = '+'  // 简单字符串
	ERROR         = '-'  // 错误
	INTEGER       = ':'  // 整数
	BULK_STRING   = '$'  // 批量字符串
	ARRAY         = '*' // 多批量字符串

	// 空值
	NULL          = "$-1\r\n"       // 空批量字符串
	NULL_ARRAY    = "*-1\r\n"      // 空数组
	EMPTY_STRING  = "$0\r\n\r\n"   // 空字符串
)

// 常用回复
var (
	OK_REPLY      = "+OK\r\n"
	PONG_REPLY    = "+PONG\r\n"
	QUEUED_REPLY  = "+QUEUED\r\n"
	NULL_REPLY    = "$-1\r\n"
	EMPTY_REPLY   = "*0\r\n"
)
