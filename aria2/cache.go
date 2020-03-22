package aria2

import (
	"sync"

	"github.com/coocood/freecache"
)

const cacheSize = 1 * 1024 * 1024 // 1MB is enough to cache download links(pre-allocated space)
const expireSeconds = 300         // 5min to confirm downloading

type linkCache struct {
	mu    sync.Mutex
	cache *freecache.Cache
}

func newLinkCache() *linkCache {
	c := new(linkCache)
	c.cache = freecache.NewCache(cacheSize)
	return c
}

func (c *linkCache) GetAndDel(msgID int) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, err := c.cache.GetInt(int64(msgID))
	if err != nil {
		return "", err
	}
	c.cache.DelInt(int64(msgID))
	return string(val), nil
}

func (c *linkCache) Set(msgID int, link string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cache.SetInt(int64(msgID), []byte(link), expireSeconds)
}

func (c *linkCache) Del(msgID int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cache.DelInt(int64(msgID))
}
