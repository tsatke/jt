package jar

import (
	"io"

	lru "github.com/hashicorp/golang-lru"
)

type Cache struct {
	lru *lru.Cache
}

func NewCache(size int) (*Cache, error) {
	lru, err := lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		_ = value.(io.Closer).Close()
	})
	if err != nil {
		return nil, err
	}
	return &Cache{
		lru: lru,
	}, nil
}

func (c *Cache) Add(k string, f *File) {
	c.lru.Add(k, f)
}

func (c *Cache) Get(k string) (*File, bool) {
	v, ok := c.lru.Get(k)
	if !ok {
		return nil, false
	}
	return v.(*File), true
}

func (c *Cache) Close() error {
	c.lru.Purge()
	return nil
}
