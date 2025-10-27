package database

import (
	"go-redis/datastruct/dict"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
	"strings"
)

type DB struct {
	index int
	data  dict.Dict
}

type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine [][]byte

func makeDB() *DB {
	return &DB{
		data: dict.MakeSyncDict(),
	}
}

func (db *DB) Exec(conn resp.Connection, cmdLine CmdLine) resp.Reply {
	// PING SET SETNX
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "'")
	}
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}

	fun := cmd.exector
	return fun(db, cmdLine[1:])
}

// 定长指令 SET K V -> arity = 3
// 变长指令 EXISTS k1 k2 arity = -2  负数的绝对值表示最少的参数个数
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return arity == argNum
	}
	return argNum >= -arity
}

func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	value, exists := db.data.Get(key)
	return value.(*database.DataEntity), exists
}

func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		if _, exists := db.data.Get(key); exists {
			db.data.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (db *DB) Flush() {
	db.data.Clear()
}

func (db *DB) Keys() []string {
	return db.data.Keys()
}
