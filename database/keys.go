package database

import (
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/lib/wildcard"
	"go-redis/resp/reply"
)

// DEL
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}
	deleted := db.Removes(keys...)

	if deleted > 0 {
		db.AddAof(utils.ToCmdLine2("del", args...))
	}
	return reply.MakeIntReply(int64(deleted))
}

// EXISTS
func execExists(db *DB, args [][]byte) resp.Reply {
	result := 0
	for _, arg := range args {
		if _, ok := db.GetEntity(string(arg)); ok {
			result++
		}
	}
	return reply.MakeIntReply(int64(result))
}

// KEYS * 列出所有的key
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})

	return reply.MakeMultiBulkReply(result)
}

// FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()

	db.AddAof(utils.ToCmdLine2("flushdb"))
	return reply.MakeOkReply()
}

// TYPE K1
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeStatusReply("none")
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
		// TODO：实现其他数据结构时需要提供实现
	}
	return reply.MakeUnKnowErrReply()
}

// RENAME k1 k2
func execRename(db *DB, args [][]byte) resp.Reply {
	key1 := string(args[0])
	key2 := string(args[1])
	entity, exists := db.GetEntity(key1)
	if !exists {
		return reply.MakeErrReply("no such key")
	}
	db.Remove(key1)
	db.PutEntity(key2, entity)

	db.AddAof(utils.ToCmdLine2("rename", args...))

	return reply.MakeOkReply()
}

// RENAMENX
func execRenameNX(db *DB, args [][]byte) resp.Reply {
	key1 := string(args[0])
	key2 := string(args[1])

	if _, ok := db.GetEntity(key2); ok {
		return reply.MakeIntReply(0)
	}

	entity, exists := db.GetEntity(key1)
	if !exists {
		return reply.MakeErrReply("no such key")
	}
	db.Remove(key1)
	db.PutEntity(key2, entity)

	db.AddAof(utils.ToCmdLine2("renamenx", args...))
	
	return reply.MakeIntReply(1)
}

func init() {
	RegisterCommand("del", execDel, -2)
	RegisterCommand("exists", execExists, -2)
	RegisterCommand("keys", execKeys, 2)
	RegisterCommand("flushdb", execFlushDB, -1)
	RegisterCommand("type", execType, 2)
	RegisterCommand("rename", execRename, 3)
	RegisterCommand("renameNX", execRenameNX, 3)
}
