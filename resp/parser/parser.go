package parser

import (
	"bufio"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

type Payload struct {
	Data resp.Reply
	Err  error
}

type readState struct {
	readingMultiLine  bool     // 解析器正在解析单行/多行数据；一个\r\n代表一行
	expectedArgsCount int      // 正在读取的指令应该有几个参数
	msgType           byte     // 用户指令类型
	args              [][]byte // 用户传递的具体指令
	bulkLen           int64    // 预设读取数据的长度
}

func (state *readState) finished() bool {
	return state.expectedArgsCount > 0 && len(state.args) == state.expectedArgsCount
}
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()

	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		logger.Info("Read line:", string(msg))
		if err != nil {
			if ioErr {
				ch <- &Payload{nil, err}
				close(ch)
				return
			}
			ch <- &Payload{nil, err}
			state = readState{}
			continue
		}

		// 判断是否是多行解析模式
		if !state.readingMultiLine {
			if msg[0] == '*' {
				// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{nil, err}
					state = readState{}
					continue
				}
				// *0\r\n
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' {
				logger.Info("收到多行字符串：", string(msg))
				// $4\r\nPING\r\n
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{nil, err}
					state = readState{}
					continue
				}

				//$-1\r\n
				if state.bulkLen == -1 {
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{}
					continue
				}
			} else {
				lineReply, err := parseSingleLineReply(msg)
				ch <- &Payload{
					Data: lineReply,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else {
			// $3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
			// PING\r\n
			err := readBody(msg, &state)
			if err != nil {
				ch <- &Payload{nil, err}
				state = readState{}
				continue
			}
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{result, err}
				state = readState{}
				continue
			}
		}
	}
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// $4\r\nPING\r\n
// 返回值 (读取的数据， 是否有io错误， error)
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var msg []byte
	var err error
	if state.bulkLen == 0 {
		// 1. 没有 $, 可以按照\r\n切分
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else {
		// 2，有$ 严格按照$后面的数字读取字符个数
		msg = make([]byte, state.bulkLen+2)
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}

	return msg, false, nil
}

// *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// 如果开始是 '*' 将解析器设置为多行模式 设置解析参数个数
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, state.expectedArgsCount)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// $3\r\nSET\r\n
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// +OK -err :5\r\n
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// $3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// PING\r\n
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2]
	var err error
	// $3
	if line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		// $0\r\n
		if state.bulkLen <= 0 {
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		} else {

		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
