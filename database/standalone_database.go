package database

import (
	"go-redis/aof"
	"go-redis/config"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strconv"
	"strings"
)

type StandaloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

func NewStandaloneDatabase() *StandaloneDatabase {
	database := &StandaloneDatabase{}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		db := makeDB()
		db.index = i
		database.dbSet[i] = db
	}
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
		for _, db := range database.dbSet {
			sdb := db
			sdb.AddAof = func(cmdLine CmdLine) {
				database.aofHandler.AddAof(sdb.index, cmdLine)
			}
		}
	}
	return database
}

func (d *StandaloneDatabase) Exec(client resp.Connection, args database.CmdLine) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()

	cmdName := strings.ToLower(string(args[0]))
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, d, args[1:])
	}

	index := client.GetDBIndex()
	if index < 0 || index >= len(d.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}
	db := d.dbSet[index]
	return db.Exec(client, args)
}

func (d *StandaloneDatabase) Close() {

}

func (d *StandaloneDatabase) AfterClientClose(client resp.Connection) {

}

// select 4
func execSelect(conn resp.Connection, database *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex < 0 || dbIndex >= len(database.dbSet) {
		return reply.MakeErrReply("ERR DB index is out of range")
	}

	conn.SelectDB(dbIndex)
	return reply.MakeOkReply()
}
