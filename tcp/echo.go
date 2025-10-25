package tcp

import (
	"bufio"
	"context"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

type EchoClient struct {
	Conn    net.Conn
	Waiting wait.Wait
}

func (client *EchoClient) Close() error {
	client.Waiting.WaitWithTimeout(10 * time.Second)
	client.Conn.Close()
	return nil
}

type EchoHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

func MakeEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if handler.closing.Get() {
		conn.Close()
	}
	client := &EchoClient{
		Conn: conn,
	}
	handler.activeConn.Store(client, struct{}{})
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Info("Connection closed.")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		logger.Info("Message received: ", msg)

		client.Waiting.Add(1)
		b := []byte(msg)
		conn.Write(b)
		client.Waiting.Done()
	}
	handler.activeConn.Delete(client)
}

func (handler *EchoHandler) Close() error {
	logger.Info("Closing connection")
	handler.closing.Set(true)
	handler.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*EchoClient)
		client.Close()
		return true
	})
	return nil
}
