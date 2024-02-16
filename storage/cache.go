package storage

import "time"

type cache struct {
	stor Storage
}

func NewCache(stor Storage, size int, ttl time.Duration) Storage {
	return
}
