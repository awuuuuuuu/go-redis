package database

import (
	"go-redis/config"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strconv"
	"strings"
)

type Database struct {
	dbSet []*DB
}

func NewDatabase() *Database {
	database := &Database{make([]*DB, config.Properties.Databases)}
	for i := range database.dbSet {
		db := makeDB()
		db.index = i
		database.dbSet[i] = db
	}
	return database
}

func (d *Database) Exec(client resp.Connection, args database.CmdLine) resp.Reply {
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

func (d *Database) Close() {

}

func (d *Database) AfterClientClose(client resp.Connection) {

}

// select 4
func execSelect(conn resp.Connection, database *Database, args [][]byte) resp.Reply {
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
