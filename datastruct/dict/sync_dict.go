package dict

import "sync"

type SyncDict struct {
	m sync.Map
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

func (dict *SyncDict) Get(key string) (value interface{}, exists bool) {
	value, ok := dict.m.Load(key)
	return value, ok
}

func (dict *SyncDict) Len() int {
	length := 0
	dict.m.Range(func(key, value interface{}) bool {
		length++
		return true
	})
	return length
}

func (dict *SyncDict) Put(key string, value interface{}) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, value)
	if existed {
		return 0
	}
	return 1
}

func (dict *SyncDict) PutIfAbsent(key string, value interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		return 0
	}
	dict.m.Store(key, value)
	return 1
}

func (dict *SyncDict) PutIfExists(key string, value interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if !existed {
		return 0
	}
	dict.m.Store(key, value)
	return 1
}

func (dict *SyncDict) Remove(key string) {
	_, existed := dict.m.Load(key)
	if existed {
		dict.m.Delete(key)
	}
}

func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

func (dict *SyncDict) Keys() []string {
	result := make([]string, dict.Len())
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		return true
	})
	return result
}

func (dict *SyncDict) RandomKeys(limit int) []string {
	result := make([]string, limit)
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value interface{}) bool {
			result[i] = key.(string)
			return false
		})
	}
	return result
}

func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	result := make([]string, limit)
	i := 0
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		return i < limit
	})
	return result
}

func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict()
}
