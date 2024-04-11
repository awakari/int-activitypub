package mastodon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/util"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Service interface {
	SearchAndAdd(ctx context.Context, subId, q string, limitPerTerm uint32) (n uint32, err error)
}

type service struct {
	clientHttp  *http.Client
	userAgent   string
	cfgMastodon config.MastodonConfig
}

const limitRespBodyLen = 1_048_576
const minWordLen = 3

func NewService(clientHttp *http.Client, userAgent string, cfgMastodon config.MastodonConfig) Service {
	return service{
		clientHttp:  clientHttp,
		userAgent:   userAgent,
		cfgMastodon: cfgMastodon,
	}
}

func (svc service) SearchAndAdd(ctx context.Context, subId, q string, limitPerTerm uint32) (n uint32, err error) {
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
		tn, tErr = svc.searchAndAdd(ctx, subId, t, limitPerTerm)
		n += tn
		err = errors.Join(err, tErr)
	}
	return
}

func (svc service) searchAndAdd(ctx context.Context, subId, term string, limit uint32) (n uint32, err error) {
	reqQuery := "?q=" + term + "&type=statuses&limit=" + strconv.Itoa(int(limit))
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, svc.cfgMastodon.Endpoint+reqQuery, nil)
	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+svc.cfgMastodon.Client.Token)
		req.Header.Add("User-Agent", svc.userAgent)
		resp, err = svc.clientHttp.Do(req)
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
			if !s.Sensitive {
				fmt.Printf("Add new source by the subscription: %s, query: %s\n", s.Account.Uri, term)
			}
		}
	}
	return
}
