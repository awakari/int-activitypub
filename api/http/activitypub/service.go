package activitypub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"github.com/superseriousbusiness/httpsig"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Service interface {
	FetchActor(ctx context.Context, self vocab.IRI) (a vocab.Actor, err error)
	SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI) (err error)
}

type service struct {
	clientHttp *http.Client
	userAgent  string
	privKey    []byte
}

const limitRespBodyLen = 65_536

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

func NewService(clientHttp *http.Client, userAgent string, privKey []byte) Service {
	return service{
		clientHttp: clientHttp,
		userAgent:  userAgent,
		privKey:    privKey,
	}
}

func (svc service) FetchActor(ctx context.Context, addr vocab.IRI) (actor vocab.Actor, err error) {
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
		req.Header.Add("User-Agent", svc.userAgent)
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
	//
	if err != nil {
		err = fmt.Errorf("%w %s: %s", ErrActorFetch, addr, err)
	}
	return
}

func (svc service) SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI) (err error) {
	//
	var d []byte
	d, err = json.Marshal(a)
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
		req.Header.Add("User-Agent", svc.userAgent)
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
	if err == nil {
		signer, _, err = httpsig.NewSigner(prefs, digestAlgorithm, headersToSign, httpsig.Signature, 120)
	}
	var privKey any
	if err == nil {
		privKey, err = ssh.ParseRawPrivateKey(svc.privKey)
		if err != nil {
			err = fmt.Errorf("failed to parse the private key: %w", err)
		}
	}
	if err == nil {
		err = signer.SignRequest(privKey, fmt.Sprintf("https://%s/actor#main-key", svc.userAgent), req, data)
		if err != nil {
			err = fmt.Errorf("failed to sign the follow request: %w", err)
		}
	}
	return
}
