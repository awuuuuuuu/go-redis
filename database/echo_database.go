package database

import (
	databaseface "go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

func (e EchoDatabase) Exec(client resp.Connection, args databaseface.CmdLine) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

func (e EchoDatabase) Close() {
}

func (e EchoDatabase) AfterClientClose(client resp.Connection) {
}
