package service

import (
	"context"
	"errors"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/storage"
)

type Service interface {
	Follow(ctx context.Context, addr string) (err error)
	Read(ctx context.Context, addr string) (a model.Actor, err error)
	List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error)
	Unfollow(ctx context.Context, addr string) (err error)
}

var ErrInternal = errors.New("internal failure")
var ErrInvalid = errors.New("invalid argument")
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("already exists")

type service struct {
	stor storage.Storage
}

func NewService(stor storage.Storage) Service {
	return service{
		stor: stor,
	}
}

func (s service) Follow(ctx context.Context, addr string) (err error) {
	//TODO implement me
	panic("implement me")
}

func (s service) Read(ctx context.Context, addr string) (a model.Actor, err error) {
	//TODO implement me
	panic("implement me")
}

func (s service) List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error) {
	//TODO implement me
	panic("implement me")
}

func (s service) Unfollow(ctx context.Context, addr string) (err error) {
	//TODO implement me
	panic("implement me")
}
