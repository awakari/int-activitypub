package activitypub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/util"
	vocab "github.com/go-ap/activitypub"
	apiPromV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	modelProm "github.com/prometheus/common/model"
	"github.com/superseriousbusiness/httpsig"
	"github.com/writeas/go-nodeinfo"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Service interface {
	ResolveActorLink(ctx context.Context, host, name string) (self vocab.IRI, err error)
	FetchActor(ctx context.Context, addr vocab.IRI) (a vocab.Actor, tags util.ObjectTags, err error)
	SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI) (err error)
	nodeinfo.Resolver
}

type service struct {
	clientHttp *http.Client
	hostname   string
	privKey    []byte
	apiProm    apiPromV1.API
}

const limitRespBodyLen = 65_536
const metricQuerySubscribers = "sum by (service) (awk_subscribers_total)"

var prefs = []httpsig.Algorithm{
	httpsig.RSA_SHA256,
}
var digestAlgorithm = httpsig.DigestSha256
var headersToSign = []string{
	httpsig.RequestTarget,
	"host",
	"date",
	"digest",
}

var ErrActorWebFinger = errors.New("failed to get the webfinger data for actor")
var ErrActorFetch = errors.New("failed to get the actor")
var ErrActivitySend = errors.New("failed to send activity")

func NewService(clientHttp *http.Client, hostname string, privKey []byte, apiProm apiPromV1.API) Service {
	return service{
		clientHttp: clientHttp,
		hostname:   hostname,
		privKey:    privKey,
		apiProm:    apiProm,
	}
}

func (svc service) ResolveActorLink(ctx context.Context, host, name string) (self vocab.IRI, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(model.WebFingerFmtUrl, host, name, host), nil)
	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("User-Agent", svc.hostname)
		resp, err = svc.clientHttp.Do(req)
	}
	var data []byte
	if err == nil {
		data, err = io.ReadAll(io.LimitReader(resp.Body, limitRespBodyLen))
	}
	var wf apiHttp.WebFinger
	if err == nil {
		err = json.Unmarshal(data, &wf)
	}
	if err == nil {
		for _, l := range wf.Links {
			if l.Rel == "self" {
				self = vocab.IRI(l.Href)
				break
			}
		}
	}
	//
	if err != nil {
		err = fmt.Errorf("%w %s@%s: %s", ErrActorWebFinger, name, host, err)
	}
	return
}

func (svc service) FetchActor(ctx context.Context, addr vocab.IRI) (actor vocab.Actor, tags util.ObjectTags, err error) {
	//
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, string(addr), nil)
	var resp *http.Response
	var reqUrl *url.URL
	if err == nil {
		reqUrl, err = addr.URL()
	}
	if err == nil {
		req.Header.Set("Host", reqUrl.Host)
		req.Header.Add("Accept", "application/activity+json")
		req.Header.Add("Accept-Charset", "utf-8")
		req.Header.Add("User-Agent", svc.hostname)
		now := time.Now().UTC()
		req.Header.Set("Date", now.Format(http.TimeFormat))
	}
	//
	err = svc.signRequest(req, []byte{})
	//
	if err == nil {
		resp, err = svc.clientHttp.Do(req)
	}
	var data []byte
	if err == nil {
		data, err = io.ReadAll(io.LimitReader(resp.Body, limitRespBodyLen))
	}
	if err == nil && resp.StatusCode > 299 {
		err = fmt.Errorf("%w %s: response status %d, message: %s", ErrActorFetch, addr, resp.StatusCode, string(data))
	}
	if err == nil {
		err = json.Unmarshal(data, &actor)
	}
	if err == nil {
		err = json.Unmarshal(data, &tags)
	}
	//
	if err != nil {
		err = fmt.Errorf("%w %s: %s", ErrActorFetch, addr, err)
	}
	return
}

func (svc service) SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI) (err error) {
	//
	aFixed, _ := apiHttp.FixContext(a)
	var d []byte
	d, err = json.Marshal(aFixed)
	var req *http.Request
	if err == nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, string(inbox), bytes.NewReader(d))
	}
	var inboxUrl *url.URL
	if err == nil {
		inboxUrl, err = inbox.URL()
	}
	if err == nil {
		req.Header.Set("Host", inboxUrl.Host)
		req.Header.Set("Content-Type", "application/ld+json; profile=\"http://www.w3.org/ns/activitystreams\"")
		req.Header.Add("Accept-Charset", "utf-8")
		req.Header.Add("User-Agent", svc.hostname)
		now := time.Now().UTC()
		req.Header.Set("Date", now.Format(http.TimeFormat))
	}
	//
	err = svc.signRequest(req, d)
	//
	var resp *http.Response
	if err == nil {
		resp, err = svc.clientHttp.Do(req)
	}
	var respData []byte
	if err == nil {
		respData, err = io.ReadAll(io.LimitReader(resp.Body, limitRespBodyLen))
		if err == nil && resp.StatusCode >= 300 {
			err = fmt.Errorf("follow response status: %d, headers: %+v, content:\n%s\n", resp.StatusCode, resp.Header, string(respData))
		}
	}
	//
	if err != nil {
		err = fmt.Errorf("%w: %s", ErrActivitySend, err)
	}
	return
}

func (svc service) signRequest(req *http.Request, data []byte) (err error) {
	var signer httpsig.Signer
	signer, _, err = httpsig.NewSigner(prefs, digestAlgorithm, headersToSign, httpsig.Signature, 120)
	var privKey any
	if err == nil {
		privKey, err = ssh.ParseRawPrivateKey(svc.privKey)
		if err != nil {
			err = fmt.Errorf("failed to parse the private key: %w", err)
		}
	}
	if err == nil {
		err = signer.SignRequest(privKey, fmt.Sprintf("https://%s/actor#main-key", svc.hostname), req, data)
		if err != nil {
			err = fmt.Errorf("failed to sign the follow request: %w", err)
		}
	}
	return
}

func (svc service) IsOpenRegistration() (bool, error) {
	return true, nil
}

func (svc service) Usage() (u nodeinfo.Usage, err error) {
	ctx := context.TODO()
	u.Users.Total, err = svc.getMetricInt(ctx, metricQuerySubscribers, 0)
	u.Users.ActiveMonth = u.Users.Total
	u.Users.ActiveHalfYear = u.Users.Total
	return
}

func (svc service) getMetricInt(ctx context.Context, q string, d time.Duration) (num int, err error) {
	var t time.Time
	switch d {
	case 0:
		t = time.Now().UTC()
	default:
		t = time.Now().UTC().Add(-d)
	}
	var v modelProm.Value
	v, _, err = svc.apiProm.Query(ctx, q, t)
	if err == nil {
		if v.Type() == modelProm.ValVector {
			if vv := v.(modelProm.Vector); len(vv) > 0 {
				num = int(vv[0].Value)
			}
		}
	}
	return
}
