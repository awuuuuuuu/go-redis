package cluster

import (
	"context"
	"errors"
	"go-redis/resp/client"

	pool "github.com/jolestar/go-commons-pool"
)

type connectionFactory struct {
	Peer string
}

func (c *connectionFactory) MakeObject(ctx context.Context) (*pool.PooledObject, error) {
	client, err := client.MakeClient(c.Peer)
	if err != nil {
		return nil, err
	}
	client.Start()
	return pool.NewPooledObject(client), nil
}

func (c *connectionFactory) DestroyObject(ctx context.Context, object *pool.PooledObject) error {
	client, ok := object.Object.(*client.Client)
	if !ok {
		return errors.New("object is not a client")
	}
	client.Close()
	return nil
}

func (c *connectionFactory) ValidateObject(ctx context.Context, object *pool.PooledObject) bool {
	return true
}

func (c *connectionFactory) ActivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}

func (c *connectionFactory) PassivateObject(ctx context.Context, object *pool.PooledObject) error {
	return nil
}
