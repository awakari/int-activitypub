package mastodon

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/awakari/int-activitypub/config"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Service interface {
	SearchAndAdd(ctx context.Context, subId, q string, limit uint32) (n uint32, err error)
}

type service struct {
	clientHttp  *http.Client
	userAgent   string
	cfgMastodon config.MastodonConfig
}

const limitRespBodyLen = 262_144
const limitSearchResults = 16

func NewService(clientHttp *http.Client, userAgent string, cfgMastodon config.MastodonConfig) Service {
	return service{
		clientHttp:  clientHttp,
		userAgent:   userAgent,
		cfgMastodon: cfgMastodon,
	}
}

func (svc service) SearchAndAdd(ctx context.Context, subId, q string, limit uint32) (n uint32, err error) {
	reqQuery := strings.Replace(q, " ", "+", -1)
	reqQuery = "?q=" + url.QueryEscape(q) + "&type=statuses&limit=" + strconv.Itoa(limitSearchResults)
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, svc.cfgMastodon.Endpoint+reqQuery, nil)
	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", "application/json")
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
				fmt.Printf("Add new source by the subscription: %s, query: %s\n", s.Account.Uri, q)
			}
		}
	}
	return
}
