package database

import "strings"

var cmdTable = make(map[string]*command)

type command struct {
	exector ExecFunc
	arity   int // 参数数量
}

func RegisterCommand(name string, ex ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{exector: ex, arity: arity}
}
