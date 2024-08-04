package writer

import (
	"context"
	"errors"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	"github.com/awakari/client-sdk-go/api/grpc/limits"
	"github.com/awakari/client-sdk-go/api/grpc/permits"
	"github.com/awakari/client-sdk-go/api/grpc/resolver"
	"github.com/awakari/client-sdk-go/model"
	"github.com/cenkalti/backoff/v4"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"google.golang.org/grpc/metadata"
	"io"
	"log/slog"
	"sync"
	"time"
)

type Service interface {
	io.Closer
	Write(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error)
}

type service struct {
	cache            *expirable.LRU[string, model.Writer[*pb.CloudEvent]]
	cacheLock        *sync.Mutex
	clientAwk        api.Client
	backoffTimeLimit time.Duration
	log              *slog.Logger
}

const accSep = ":"
const backoffInitDelay = 100 * time.Millisecond
const cacheWriterSize = 1024
const cacheWriterTtl = 24 * time.Hour

var ErrWrite = errors.New("failed to write event")
var errNoAck = errors.New("event is not accepted")

func NewService(clientAwk api.Client, backoffTimeLimit time.Duration, log *slog.Logger) Service {
	funcEvict := func(_ string, w model.Writer[*pb.CloudEvent]) {
		_ = w.Close()
	}
	return service{
		cache:            expirable.NewLRU[string, model.Writer[*pb.CloudEvent]](cacheWriterSize, funcEvict, cacheWriterTtl),
		cacheLock:        &sync.Mutex{},
		clientAwk:        clientAwk,
		backoffTimeLimit: backoffTimeLimit,
		log:              log,
	}
}

func (svc service) Close() (err error) {
	svc.cacheLock.Lock()
	defer svc.cacheLock.Unlock()
	for _, k := range svc.cache.Keys() {
		w, found := svc.cache.Get(k)
		if found {
			err = errors.Join(err, w.Close())
		}
	}
	svc.cache.Purge()
	return
}

func (svc service) Write(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error) {
	err = svc.getWriterAndPublish(ctx, evt, groupId, userId)
	if err != nil {
		err = svc.retryBackoff(func() error {
			return svc.getWriterAndPublish(ctx, evt, groupId, userId)
		})
	}
	if err != nil {
		err = fmt.Errorf("%w id: %s, cause: %s", ErrWrite, evt.Id, err)
	}
	return
}

func (svc service) getWriterAndPublish(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error) {
	var w model.Writer[*pb.CloudEvent]
	w, err = svc.getWriter(ctx, groupId, userId)
	if err == nil {
		err = svc.publish(w, evt)
		switch {
		case errors.Is(err, limits.ErrReached):
			svc.log.Debug(fmt.Sprintf("Publish failure: evt.Id=%s, userId=%s, err=%s", evt.Id, userId, err))
			err = nil   // don't retry this time
			fallthrough // reopen the writer the next time
		case errors.Is(err, limits.ErrUnavailable):
			fallthrough
		case errors.Is(err, permits.ErrUnavailable):
			fallthrough
		case errors.Is(err, resolver.ErrUnavailable):
			fallthrough
		case errors.Is(err, resolver.ErrInternal):
			fallthrough
		case errors.Is(err, io.EOF):
			// close and remove the writer from the cache
			k := writerKey(groupId, userId)
			svc.cacheLock.Lock()
			defer svc.cacheLock.Unlock()
			svc.cache.Remove(k)
			_ = w.Close()
		}
	}
	return
}

func (svc service) getWriter(ctx context.Context, groupId, userId string) (w model.Writer[*pb.CloudEvent], err error) {
	k := writerKey(groupId, userId)
	svc.cacheLock.Lock()
	defer svc.cacheLock.Unlock()
	w, found := svc.cache.Get(k)
	if !found {
		ctxGroupId := metadata.AppendToOutgoingContext(ctx, "x-awakari-group-id", groupId)
		w, err = svc.clientAwk.OpenMessagesWriter(ctxGroupId, userId)
		if err == nil {
			svc.cache.Add(k, w)
		}
	}
	return
}

func (svc service) publish(w model.Writer[*pb.CloudEvent], evt *pb.CloudEvent) (err error) {
	err = svc.tryPublish(w, evt)
	if err == errNoAck {
		err = svc.retryBackoff(func() error {
			return svc.tryPublish(w, evt)
		})
	}
	return
}

func (svc service) tryPublish(w model.Writer[*pb.CloudEvent], evt *pb.CloudEvent) (err error) {
	var ackCount uint32
	ackCount, err = w.WriteBatch([]*pb.CloudEvent{evt})
	if err == nil && ackCount < 1 {
		err = errNoAck //  error to retry w/o reopening the writer
	}
	return
}

func (svc service) retryBackoff(op func() error) (err error) {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = backoffInitDelay
	b.MaxElapsedTime = svc.backoffTimeLimit
	err = backoff.Retry(op, b)
	return
}

func writerKey(groupId, userId string) (k string) {
	k = fmt.Sprintf("%s%s%s", groupId, accSep, userId)
	return
}
