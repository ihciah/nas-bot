package aria2

import (
	"bytes"
	"encoding/gob"
	"errors"
	"strconv"
	"sync"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const (
	defaultExpiration = 5 * time.Minute
	cleanupInterval   = 10 * time.Minute
)

type linkFileUnion struct {
	Link     string
	FileName string
	FileData []byte
}

func (u *linkFileUnion) IsLink() bool {
	return len(u.Link) != 0
}

func (u *linkFileUnion) IsFile() bool {
	return len(u.FileData) != 0
}

type cache struct {
	mu    sync.Mutex
	cache *gocache.Cache
}

func newCache() *cache {
	c := new(cache)
	c.cache = gocache.New(defaultExpiration, cleanupInterval)
	return c
}

func (c *cache) GetAndDel(msgID int) (linkFileUnion, error) {
	var union linkFileUnion
	val, ok := c.cache.Get(strconv.Itoa(msgID))
	if !ok {
		return union, errors.New("unable to get key")
	}
	b, ok := val.([]byte)
	if !ok {
		return union, errors.New("unable to get key")
	}
	c.cache.Delete(strconv.Itoa(msgID))
	err := gob.NewDecoder(bytes.NewReader(b)).Decode(&union)
	// Gob is actually useless here, using []byte is meant for replacing cache
	// easily.
	return union, err
}

func (c *cache) SetLink(msgID int, link string) error {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(linkFileUnion{Link: link})
	if err != nil {
		return err
	}
	c.cache.Set(strconv.Itoa(msgID), b.Bytes(), defaultExpiration)
	return nil
}

func (c *cache) SetFile(msgID int, fileName string, fileData []byte) error {
	var b bytes.Buffer
	err := gob.NewEncoder(&b).Encode(linkFileUnion{FileName: fileName, FileData: fileData})
	if err != nil {
		return err
	}
	c.cache.Set(strconv.Itoa(msgID), b.Bytes(), defaultExpiration)
	return nil
}

func (c *cache) Del(msgID int) {
	c.cache.Delete(strconv.Itoa(msgID))
}
