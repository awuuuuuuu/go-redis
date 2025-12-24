package cluster

import (
	"context"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/client"
	"go-redis/resp/reply"
	"strconv"
)

func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	pool, ok := cluster.peerConnectionPool[peer]
	if !ok {
		return nil, errors.New("connection not found")
	}
	object, err := pool.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	client2, ok := object.(*client.Client)
	if !ok {
		return nil, errors.New("wrong object type")
	}
	return client2, nil
}

func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	pool, ok := cluster.peerConnectionPool[peer]
	if !ok {
		return errors.New("connection not found")
	}
	err := pool.ReturnObject(context.Background(), peerClient)
	return err
}

// 转发
func (cluster *ClusterDatabase) relay(peer string, conn resp.Connection, args [][]byte) resp.Reply {
	if peer == cluster.self {
		return cluster.db.Exec(conn, args)
	}

	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeErrReply("获取兄弟结点失败")
	}
	defer func(cluster *ClusterDatabase, peer string, peerClient *client.Client) {
		cluster.returnPeerClient(peer, peerClient)
	}(cluster, peer, peerClient)

	peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(conn.GetDBIndex())))
	return peerClient.Send(args)
}

func (cluster *ClusterDatabase) broadcast(conn resp.Connection, args [][]byte) map[string]resp.Reply {
	results := make(map[string]resp.Reply)
	for _, node := range cluster.nodes {
		resp := cluster.relay(node, conn, args)
		results[node] = resp
	}
	return results
}
