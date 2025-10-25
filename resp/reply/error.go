package reply

type UnKnowErrReply struct {
}

var unKnowErrBytes = []byte("-Err unknown\r\n")

func (u *UnKnowErrReply) Error() string {
	return "Err unknown"
}

func (u *UnKnowErrReply) ToBytes() []byte {
	return unKnowErrBytes
}

var theUnKnowErrReply = &UnKnowErrReply{}

func MakeUnKnowErrReply() *UnKnowErrReply {
	return theUnKnowErrReply
}

type ArgNumErrReply struct {
	Cmd string
}

func (a *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + a.Cmd + "' command"
}

func (a *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + a.Cmd + "' command\r\n")
}

func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{cmd}
}

type SyntaxErrReply struct {
}

var syntaxErrBytes = []byte("-Err syntax error\r\n")

func (s *SyntaxErrReply) Error() string {
	return "Syntax Error"
}

func (s *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

var theSyntaxErrReply = &SyntaxErrReply{}

func MakeSyntaxErrReply(cmd string) *SyntaxErrReply {
	return theSyntaxErrReply
}

type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

func (w *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}
func (w *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

var theWrongTypeErrReply = &WrongTypeErrReply{}

func MakeWrongTypeErrReply() *WrongTypeErrReply {
	return theWrongTypeErrReply
}

type ProtocolErrReply struct {
	Msg string
}

func (p *ProtocolErrReply) Error() string {
	return "ERR Protocol Error: '" + p.Msg + "'"
}
func (p *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol Error: '" + p.Msg + "'")
}
func MakeProtocolErrReply(msg string) *ProtocolErrReply {
	return &ProtocolErrReply{msg}
}
