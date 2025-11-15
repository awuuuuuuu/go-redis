package aof

import (
	"go-redis/config"
	databaseface "go-redis/interface/database"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"os"
	"strconv"
)

// CmdLine is alias for [][]byte, represents a command line
type CmdLine = [][]byte

const aofBufferSize = 1 << 16

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

// AofHandler receive msgs from channel and write to AOF file
type AofHandler struct {
	database    databaseface.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	currentDB   int
}

func NewAofHandler(database databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.database = database

	// 加载Aof
	handler.LoadAof()

	aofFile, err := os.OpenFile(handler.aofFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile

	// channel
	handler.aofChan = make(chan *payload, aofBufferSize)
	go func() {
		handler.HandleAof()
	}()

	return handler, nil
}

func (handler *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmdLine,
			dbIndex: dbIndex,
		}
	}
}

// HandleAof payload(set k, v) <- aofChan（落盘）
func (handler *AofHandler) HandleAof() {
	handler.currentDB = -1
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {
			bytes := reply.MakeMultiBulkReply(utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(bytes)
			if err != nil {
				logger.Error(err)
				continue
			}
			handler.currentDB = p.dbIndex
		}
		bytes := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(bytes)
		if err != nil {
			logger.Error(err)
		}
	}

}

func (handler *AofHandler) LoadAof() {
	file, err := os.Open(handler.aofFilename)
	if err != nil {
		logger.Error(err)
	}
	defer file.Close()
	ch := parser.ParseStream(file)

	fakeConn := connection.Connection{}
	fakeConn.SelectDB(handler.currentDB)
	for p := range ch {
		if p.Err != nil {
			if p.Err == io.EOF {
				break
			}
			logger.Error(p.Err)
			continue
		}

		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}

		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error(p.Err)
		}

		rep := handler.database.Exec(&fakeConn, r.Args)
		if reply.IsErrReply(rep) {
			logger.Error(rep)
		}
		
	}
}
