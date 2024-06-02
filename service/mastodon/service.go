package mastodon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/service"
	"github.com/r3labs/sse/v2"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

type Service interface {
	SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error)
	ConsumeLiveStreamPublic(ctx context.Context) (err error)
}

type mastodon struct {
	clientHttp *http.Client
	userAgent  string
	cfg        config.MastodonConfig
	svc        service.Service
}

const limitRespBodyLen = 1_048_576
const minFollowersCount = 10
const minPostCount = 10

func NewService(clientHttp *http.Client, userAgent string, cfgMastodon config.MastodonConfig, svc service.Service) Service {
	return mastodon{
		clientHttp: clientHttp,
		userAgent:  userAgent,
		cfg:        cfgMastodon,
		svc:        svc,
	}
}

func (m mastodon) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error) {
	var offset int
	for n < limit {
		reqQuery := "?q=" + url.QueryEscape(q) + "&type=statuses&offset=" + strconv.Itoa(offset) + "&limit=" + strconv.Itoa(int(limit))
		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.Endpoint.Search+reqQuery, nil)
		var resp *http.Response
		if err == nil {
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Authorization", "Bearer "+m.cfg.Client.Token)
			req.Header.Add("User-Agent", m.userAgent)
			resp, err = m.clientHttp.Do(req)
		}
		var data []byte
		if err == nil {
			data, err = io.ReadAll(io.LimitReader(resp.Body, limitRespBodyLen))
		}
		var results Results
		if err == nil {
			err = json.Unmarshal(data, &results)
		}
		if err == nil {
			countResults := len(results.Statuses)
			if countResults == 0 {
				break
			}
			offset += countResults
			for _, s := range results.Statuses {
				var errReqFollow error
				if s.Sensitive {
					errReqFollow = fmt.Errorf("found account %s skip due to sensitive flag", s.Account.Uri)
				}
				if errReqFollow == nil && s.Account.FollowersCount < minFollowersCount {
					errReqFollow = fmt.Errorf("found account %s skip due low followers count %d", s.Account.Uri, s.Account.FollowersCount)
				}
				if errReqFollow == nil && s.Account.StatusesCount < minPostCount {
					errReqFollow = fmt.Errorf("found account %s skip due low post count %d", s.Account.Uri, s.Account.StatusesCount)
				}
				if errReqFollow == nil {
					_, errReqFollow = m.svc.RequestFollow(ctx, s.Account.Uri, groupId, "", subId, q)
				}
				if errReqFollow == nil {
					n++
				}
				if n > limit {
					break
				}
				err = errors.Join(err, errReqFollow)
			}
		}
	}
	return
}

func (m mastodon) ConsumeLiveStreamPublic(ctx context.Context) (err error) {
	client := sse.NewClient(m.cfg.Endpoint.Stream)
	client.Headers["Authorization"] = "Bearer " + m.cfg.Client.Token
	client.Headers["User-Agent"] = m.userAgent
	err = client.SubscribeWithContext(ctx, "", m.consumeLiveStreamEvent)
	return
}

func (m mastodon) consumeLiveStreamEvent(evt *sse.Event) {
	fmt.Println(evt.Data)
	return
}
