package storage

import (
	"context"
	"errors"
	"github.com/awakari/int-activitypub/model"
	"io"
)

type Storage interface {
	io.Closer
	Create(ctx context.Context, addr string) (err error)
	Read(ctx context.Context, addr string) (a model.Actor, err error)
	List(ctx context.Context, filter model.ActorFilter, limit uint32, cursor string, order model.Order) (page []string, err error)
	Delete(ctx context.Context, addr string) (err error)
}

var ErrInternal = errors.New("internal failure")
var ErrConflict = errors.New("already exists")
var ErrNotFound = errors.New("not found")
