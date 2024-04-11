package mastodon

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/util"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Service interface {
	SearchAndAdd(ctx context.Context, subId, groupId, q string, limitPerTerm uint32) (n uint32, err error)
}

type mastodon struct {
	clientHttp *http.Client
	userAgent  string
	cfg        config.MastodonConfig
	svc        service.Service
}

const limitRespBodyLen = 1_048_576
const minWordLen = 3

func NewService(clientHttp *http.Client, userAgent string, cfgMastodon config.MastodonConfig, svc service.Service) Service {
	return mastodon{
		clientHttp: clientHttp,
		userAgent:  userAgent,
		cfg:        cfgMastodon,
		svc:        svc,
	}
}

func (m mastodon) SearchAndAdd(ctx context.Context, subId, groupId, q string, limitPerTerm uint32) (n uint32, err error) {
	q = util.Sanitize(q)
	terms := map[string]bool{}
	for _, t := range strings.Split(q, " ") {
		if len(t) >= minWordLen {
			terms[t] = true
		}
	}
	var tn uint32
	var tErr error
	for t, _ := range terms {
		tn, tErr = m.searchAndAdd(ctx, subId, groupId, t, limitPerTerm)
		n += tn
		err = errors.Join(err, tErr)
	}
	return
}

func (m mastodon) searchAndAdd(ctx context.Context, subId, groupId, term string, limit uint32) (n uint32, err error) {
	reqQuery := "?q=" + term + "&type=statuses&limit=" + strconv.Itoa(int(limit))
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.Endpoint+reqQuery, nil)
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
		for _, s := range results.Statuses {
			var errReqFollow error
			if !s.Sensitive {
				_, errReqFollow = m.svc.RequestFollow(ctx, s.Account.Uri, groupId, s.Account.Uri)
			}
			err = errors.Join(err, errReqFollow)
		}
	}
	return
}
