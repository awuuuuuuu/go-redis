package handler

import (
	"context"
	"errors"
	"go-redis/database"
	databaseface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"net"
	"strings"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
	db         databaseface.Database
}

func MakeHandler() *RespHandler {
	db := database.NewDatabase()
	return &RespHandler{
		db: db,
	}
}

func (r *RespHandler) closeClient(client *connection.Connection) {
	client.Close()
	r.db.AfterClientClose(client)
	r.activeConn.Delete(client)
}

func (r *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if r.closing.Get() {
		conn.Close()
	}
	client := connection.NewConnection(conn)
	r.activeConn.Store(client, struct{}{})
	ch := parser.ParseStream(conn)
	for payload := range ch {
		//error
		if payload.Err != nil {
			if errors.Is(payload.Err, io.EOF) || errors.Is(payload.Err, io.ErrUnexpectedEOF) ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				r.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}

			// protocol error
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				r.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}

		if payload.Data == nil {
			continue
		}

		//exec
		bulkReply, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Info("request data type is not MultiBulkReply")
			continue
		}
		result := r.db.Exec(client, bulkReply.Args)
		if result == nil {
			_ = client.Write(unknownErrReplyBytes)
		} else {
			_ = client.Write(result.ToBytes())
		}
	}
}

func (r *RespHandler) Close() error {
	logger.Info("close resp handler")
	r.closing.Set(true)
	r.activeConn.Range(func(key, value interface{}) bool {
		c := key.(*connection.Connection)
		c.Close()
		return true
	})
	r.db.Close()
	return nil
}
