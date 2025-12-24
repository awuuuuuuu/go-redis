package cluster

import (
	"context"
	"go-redis/config"
	database2 "go-redis/database"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/lib/consistenthash"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strings"

	pool "github.com/jolestar/go-commons-pool"
)

type ClusterDatabase struct {
	self string

	nodes              []string
	peerPicker         *consistenthash.NodeMap
	peerConnectionPool map[string]*pool.ObjectPool
	db                 database.Database
}

func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:               config.Properties.Self,
		db:                 database2.NewStandaloneDatabase(),
		peerPicker:         consistenthash.NewNodeMap(nil),
		peerConnectionPool: make(map[string]*pool.ObjectPool),
	}

	ctx := context.Background()
	nodes := make([]string, 0, len(config.Properties.Peers)+1)
	nodes = append(nodes, config.Properties.Self)
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
		cluster.peerConnectionPool[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{Peer: peer})
	}
	cluster.nodes = nodes
	cluster.peerPicker.AddNode(cluster.nodes...)

	return cluster
}

var router = makeRouter()

func (cluster *ClusterDatabase) Exec(client resp.Connection, args database.CmdLine) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			result = reply.MakeUnKnowErrReply()
		}
	}()

	cmdName := strings.ToLower(string(args[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		result = reply.MakeErrReply("not support command")
	}
	result = cmdFunc(cluster, client, args)
	return
}

func (cluster *ClusterDatabase) Close() {
	cluster.db.Close()
}

func (cluster *ClusterDatabase) AfterClientClose(client resp.Connection) {
	cluster.db.AfterClientClose(client)
}
