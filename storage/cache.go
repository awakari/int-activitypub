package storage

import (
	"context"
	"github.com/awakari/int-activitypub/model"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"time"
)

type localCache struct {
	stor  Storage
	cache *expirable.LRU[string, model.Source]
}

func NewLocalCache(stor Storage, size int, ttl time.Duration) Storage {
	c := expirable.NewLRU[string, model.Source](size, nil, ttl)
	return localCache{
		stor:  stor,
		cache: c,
	}
}

func (lc localCache) Close() error {
	lc.cache.Purge()
	return lc.stor.Close()
}

func (lc localCache) Create(ctx context.Context, src model.Source) (err error) {
	err = lc.stor.Create(ctx, src)
	if err == nil {
		lc.cache.Add(src.ActorId, src)
	}
	return
}

func (lc localCache) Read(ctx context.Context, srcId string) (src model.Source, err error) {
	var found bool
	src, found = lc.cache.Get(srcId)
	if !found {
		src, err = lc.stor.Read(ctx, srcId)
	}
	return
}

func (lc localCache) Update(ctx context.Context, src model.Source) (err error) {
	err = lc.stor.Update(ctx, src)
	if err == nil {
		lc.cache.Add(src.ActorId, src)
	}
	return
}

func (lc localCache) Delete(ctx context.Context, srcId, groupId, userId string) (err error) {
	err = lc.stor.Delete(ctx, srcId, groupId, userId)
	lc.cache.Remove(srcId)
	return
}

func (lc localCache) List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	page, err = lc.stor.List(ctx, filter, limit, cursor, order)
	return
}
