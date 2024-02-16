package storage

import (
	"context"
	"errors"
	"github.com/awakari/int-activitypub/model"
	"io"
)

type Storage interface {
	io.Closer
	Create(ctx context.Context, src model.Source) (err error)
	Read(ctx context.Context, srcId string) (src model.Source, err error)
	Update(ctx context.Context, src model.Source) (err error)
	Delete(ctx context.Context, srcId, groupId, userId string) (err error)
	List(ctx context.Context, filter model.Filter, limit uint32, cursor string, order model.Order) (page []string, err error)
}

var ErrInternal = errors.New("internal failure")
var ErrConflict = errors.New("already exists")
var ErrNotFound = errors.New("not found")
