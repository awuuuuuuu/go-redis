package cluster

import (
	"bytes"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

type CmdFunc func(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply

func makeRouter() map[string]CmdFunc {
	router := make(map[string]CmdFunc)
	router["exists"] = defaultFunc // exists k1
	router["type"] = defaultFunc   // type k1
	router["set"] = defaultFunc
	router["setnx"] = defaultFunc
	router["get"] = defaultFunc
	router["getset"] = defaultFunc
	router["ping"] = selfFunc
	router["select"] = selfFunc
	router["rename"] = renameFunc
	router["renamenx"] = renamenxFunc
	router["flushdb"] = flushdbFunc
	router["del"] = delFunc

	return router
}

// GET Key // Set K1 V1
func defaultFunc(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply {
	key := string(cmdArgs[1])
	peer := cluster.peerPicker.PickNode(key)
	return cluster.relay(peer, conn, cmdArgs)
}

func selfFunc(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply {
	return cluster.db.Exec(conn, cmdArgs)
}

// renameFunc rename k1 k2
/*
GET k1 v1
DEL k1
SET k2 v1
*/
func renameFunc(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply {
	// 参数校验
	if len(cmdArgs) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'rename' command")
	}

	oldKey := cmdArgs[1]
	newKey := cmdArgs[2]

	// 检查新旧key是否相同
	if bytes.Equal(oldKey, newKey) {
		return reply.MakeErrReply("ERR source and destination objects are the same")
	}

	// 1. 获取原key的值
	getCmd := [][]byte{[]byte("GET"), oldKey}
	result1 := defaultFunc(cluster, conn, getCmd)
	if reply.IsErrReply(result1) {
		return result1 // 直接返回错误，保留原始错误信息
	}

	bulkReply, ok := result1.(*reply.BulkReply)
	if !ok || bulkReply == nil {
		return reply.MakeErrReply("ERR no such key")
	}

	// 2. 设置新key的值
	setCmd := [][]byte{[]byte("SET"), newKey, bulkReply.Arg}
	result3 := defaultFunc(cluster, conn, setCmd)
	if reply.IsErrReply(result3) {
		return result3
	}

	// 3. 删除原key（只有在设置成功后）
	delCmd := [][]byte{[]byte("DEL"), oldKey}
	delFunc(cluster, conn, delCmd) // 忽略删除结果，因为原key可能已被其他操作删除

	return reply.MakeOkReply()
}

func renamenxFunc(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply {
	// 参数校验
	if len(cmdArgs) != 3 {
		return reply.MakeErrReply("ERR wrong number of arguments for 'renamenx' command")
	}

	oldKey := cmdArgs[1]
	newKey := cmdArgs[2]

	// 检查新旧key是否相同
	if bytes.Equal(oldKey, newKey) {
		return reply.MakeIntReply(0) // Redis 在相同key时返回 0
	}

	// 1. 检查新key是否已存在
	existsCmd := [][]byte{[]byte("EXISTS"), newKey}
	existsResult := defaultFunc(cluster, conn, existsCmd)
	if reply.IsErrReply(existsResult) {
		return existsResult
	}

	intReply, ok := existsResult.(*reply.IntReply)
	if !ok {
		return reply.MakeErrReply("ERR invalid EXISTS response")
	}

	// 如果新key已存在，返回 0
	if intReply.Code > 0 {
		return reply.MakeIntReply(0)
	}

	// 2. 获取原key的值
	getCmd := [][]byte{[]byte("GET"), oldKey}
	getResult := defaultFunc(cluster, conn, getCmd)
	if reply.IsErrReply(getResult) {
		return getResult
	}

	bulkReply, ok := getResult.(*reply.BulkReply)
	if !ok || bulkReply == nil {
		return reply.MakeErrReply("ERR no such key") // 原key不存在
	}

	// 3. 设置新key的值 - 修复了这里的错误
	setCmd := [][]byte{[]byte("SET"), newKey, bulkReply.Arg} // 使用 Value 而不是 ToBytes()
	setResult := defaultFunc(cluster, conn, setCmd)
	if reply.IsErrReply(setResult) {
		return setResult
	}

	// 4. 删除原key（只有在设置成功后）
	delCmd := [][]byte{[]byte("DEL"), oldKey}
	delFunc(cluster, conn, delCmd) // 忽略删除结果

	return reply.MakeIntReply(1) // 成功返回 1
}

func flushdbFunc(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply {
	var errReply reply.ErrorReply
	for _, r := range cluster.broadcast(conn, cmdArgs) {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply != nil {
		return reply.MakeOkReply()
	}
	return reply.MakeErrReply("ERR flushdb command failed")
}

func delFunc(cluster *ClusterDatabase, conn resp.Connection, cmdArgs [][]byte) resp.Reply {
	var tot int64
	var errReply reply.ErrorReply
	for _, r := range cluster.broadcast(conn, cmdArgs) {
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}

		intReply, ok := r.(*reply.IntReply)
		if !ok {
			errReply = r.(reply.ErrorReply)
			break
		} else {
			tot += intReply.Code
		}
	}
	if errReply != nil {
		return reply.MakeErrReply("ERR del command failed")
	}
	return reply.MakeIntReply(tot)
}
