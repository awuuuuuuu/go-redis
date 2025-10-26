package reply

import (
	"bytes"
	"go-redis/interface/resp"
	"strconv"
)

var (
	nullBulkReplyBytes = []byte("$-1")
	CRLF               = "\r\n"
)

/*
	RESP Redis Serialization Protocol
	五种回复
	--正常回复
		以 "+" 开头，以 "\r\n" 结尾
		+OK\r\n
	--错误回复
		以 "-" 开头，以 "\r\n" 结尾
		-Error message\r\n
	--整数回复
		以 ":" 开头，以 "\r\n" 结尾
		:123456\r\n
	--多行字符串
		以 "$" 开头，后面跟实际发送字节数，以 "\r\n" 结尾
		$3\r\nawu\r\n
		$0\r\n\r\n 表示一个长度为 0 的字符串
		$-1\r\n 表示空值（Null），即“键不存在”或“无结果”
		$7\r\na\r\nwu\r\n
	--数组
		以 "*" 开头，后面跟成员个数
		// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
*/

// BulkReply 字符串
type BulkReply struct {
	Arg []byte // "awu" "$3\r\nawu\r\n"
}

func (b *BulkReply) ToBytes() []byte {
	if len(b.Arg) == 0 {
		return nullBulkReplyBytes
	}
	return []byte("$" + strconv.Itoa(len(b.Arg)) + CRLF + string(b.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{arg}
}

// MultiBulkReply 数组
type MultiBulkReply struct {
	Args [][]byte
}

func (m *MultiBulkReply) ToBytes() []byte {
	argLen := len(m.Args)
	if argLen == 0 {
		return nullBulkReplyBytes
	}
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range m.Args {
		if arg == nil {
			buf.WriteString(string(nullBulkReplyBytes) + CRLF)
		}
		buf.WriteString("$" + strconv.Itoa(argLen) + CRLF + string(arg) + CRLF)
	}
	return buf.Bytes()
}
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{args}
}

type StatusReply struct {
	Status string
}

func (s *StatusReply) ToBytes() []byte {
	return []byte("+" + s.Status + CRLF)
}
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{status}
}

type IntReply struct {
	Code int64
}

func (i *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(i.Code, 64) + CRLF)
}
func MakeIntReply(code int64) *IntReply {
	return &IntReply{code}
}

type StandardErrReply struct {
	Status string
}

func (s *StandardErrReply) Error() string {
	return s.Status
}

func (s *StandardErrReply) ToBytes() []byte {
	return []byte("-" + s.Status + CRLF)
}

func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{status}
}

func IsErrReply(reply resp.Reply) bool {
	if reply.ToBytes()[0] == '-' {
		return true
	}
	return false
}

type ErrorReply interface {
	Error() string
	ToBytes() []byte
}
