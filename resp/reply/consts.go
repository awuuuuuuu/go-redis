package reply

// PongReply 回复 PONG
type PongReply struct {
}

var pongBytes = []byte("+PONG\r\n")

func (p *PongReply) ToBytes() []byte {
	return pongBytes
}

var thePongReply = PongReply{}

func MakePongReply() *PongReply {
	return &thePongReply
}

// OkReply 回复 OK
type OkReply struct {
}

var okBytes = []byte("+OK\r\n")

func (ok *OkReply) ToBytes() []byte {
	return okBytes
}

var theOkReply = &OkReply{}

func MakeOkReply() *OkReply {
	return theOkReply
}

// NullBulkReply 回复 空字符串
type NullBulkReply struct {
}

var nullBulkBytes = []byte("$-1\r\n")

func (n *NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

var theNullBulkReply = &NullBulkReply{}

func MakeNullBulkReply() *NullBulkReply {
	return theNullBulkReply
}

// EmptyMultiBulkReply 回复 空数组
type EmptyMultiBulkReply struct {
}

var emptyMultiBulkBytes = []byte("*0\r\n")

func (em *EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

var theEmptyMultiBulkReply = &EmptyMultiBulkReply{}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return theEmptyMultiBulkReply
}

// NoReply 回复 空
type NoReply struct {
}

var noBytes = []byte("")

func (n *NoReply) ToBytes() []byte {
	return noBytes
}

var theNoReply = &NoReply{}

func MakeNoReply() *NoReply {
	return theNoReply
}
