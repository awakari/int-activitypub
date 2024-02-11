package service

import "github.com/awakari/int-activitypub/storage"

type Service interface {
}

type service struct {
	stor storage.Storage
}

func NewService(stor storage.Storage) Service {
	return service{
		stor: stor,
	}
}
