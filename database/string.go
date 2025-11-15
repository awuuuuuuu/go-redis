package database

import (
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

// GET
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeBulkReply(entity.Data.([]byte))
}

// SET
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	db.PutEntity(key, &database.DataEntity{Data: value})

	db.AddAof(utils.ToCmdLine2("set", args...))

	return reply.MakeOkReply()
}

// SETNX
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := string(args[1])
	result := db.PutIfAbsent(key, &database.DataEntity{Data: value})

	db.AddAof(utils.ToCmdLine2("setnx", args...))

	return reply.MakeIntReply(int64(result))
}

// GETSET k1 v1
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := string(args[1])
	entity, exists := db.GetEntity(key)
	db.PutEntity(key, &database.DataEntity{Data: value})

	db.AddAof(utils.ToCmdLine2("getset", args...))

	if !exists {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeBulkReply(entity.Data.([]byte))
}

// STRLEN
func execStrlen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeIntReply(int64(len(entity.Data.([]byte))))
}

func init() {
	RegisterCommand("get", execGet, 2)
	RegisterCommand("set", execSet, 3)
	RegisterCommand("setnx", execSetNX, 3)
	RegisterCommand("getset", execGetSet, 3)
	RegisterCommand("strlen", execStrlen, 2)
}
