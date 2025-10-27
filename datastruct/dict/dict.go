package dict

type Consumer func(key string, value interface{}) bool

type Dict interface {
	Get(key string) (value interface{}, exists bool)
	Len() int
	Put(key string, value interface{}) (result int)
	PutIfAbsent(key string, value interface{}) (result int)
	PutIfExists(key string, value interface{}) (result int)
	Remove(key string)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string
	RandomDistinctKeys(limit int) []string
	Clear()
}
