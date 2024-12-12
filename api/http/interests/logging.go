package interests

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/model/interest"
	"github.com/awakari/int-activitypub/util"
	"log/slog"
)

type logging struct {
	svc Service
	log *slog.Logger
}

func NewLogging(svc Service, log *slog.Logger) Service {
	return logging{
		svc: svc,
		log: log,
	}
}

func (l logging) Read(ctx context.Context, groupId, userId, subId string) (subData interest.Data, err error) {
	subData, err = l.svc.Read(ctx, groupId, userId, subId)
	l.log.Log(ctx, util.LogLevel(err), fmt.Sprintf("interests.Read(%s, %s, %s): %v, %s", groupId, userId, subId, subData, err))
	return
}
