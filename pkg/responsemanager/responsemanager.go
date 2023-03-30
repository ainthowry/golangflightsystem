package responsemanager

import (
	"context"
	"encoding/binary"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/cespare/xxhash/v2"
)

type ResponseManager struct {
	manager bigcache.BigCache
}

func NewResponseManager() *ResponseManager {
	responseManager, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(5*time.Minute))

	return &ResponseManager{
		manager: *responseManager,
	}
}

func (responseManager *ResponseManager) GetHashKey(reqId uint32, user string) []byte {
	hasher := xxhash.New()
	hasher.Write([]byte(user))
	hash := hasher.Sum64()
	hash ^= uint64(reqId)

	hashKey := make([]byte, 12)
	binary.LittleEndian.PutUint64(hashKey[:8], hash)
	binary.LittleEndian.PutUint32(hashKey[8:], reqId)
	return hashKey
}

func (responseManager *ResponseManager) GetCachedResponse(hashKey []byte) ([]byte, error) {
	cachedResponse, err := responseManager.manager.Get(string(hashKey))
	if err != nil {
		return nil, err
	}

	return cachedResponse, nil
}

func (responseManager *ResponseManager) SetCachedResponse(hashKey []byte, response []byte) error {
	err := responseManager.manager.Set(string(hashKey), response)
	if err != nil {
		return err
	}

	return nil
}
