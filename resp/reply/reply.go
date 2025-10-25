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

func MakeStandardErrReply(status string) *StandardErrReply {
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
