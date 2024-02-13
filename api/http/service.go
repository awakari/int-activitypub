package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	vocab "github.com/go-ap/activitypub"
	"github.com/superseriousbusiness/httpsig"
	"golang.org/x/crypto/ssh"
	"io"
	"net/http"
	"time"
)

type Service interface {
	ResolveActorLink(ctx context.Context, host, name string) (self vocab.IRI, err error)
	RequestFollow(ctx context.Context, host string, addr vocab.IRI, inbox vocab.IRI) (err error)
	FetchActor(ctx context.Context, self vocab.IRI) (a vocab.Actor, err error)
}

type service struct {
	clientHttp *http.Client
	hostSelf   string
	privKey    []byte
}

const fmtWebFinger = "https://%s/.well-known/webfinger?resource=acct:%s@%s"
const limitRespLen = 65_536

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

func NewService(clientHttp *http.Client, hostSelf string, privKey []byte) Service {
	return service{
		clientHttp: clientHttp,
		hostSelf:   hostSelf,
		privKey:    privKey,
	}
}

func (svc service) ResolveActorLink(ctx context.Context, host, name string) (self vocab.IRI, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf(fmtWebFinger, host, name, host), nil)
	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", "application/jrd+json")
		req.Header.Add("User-Agent", svc.hostSelf)
		resp, err = svc.clientHttp.Do(req)
	}
	var data []byte
	if err == nil {
		data, err = io.ReadAll(io.LimitReader(resp.Body, limitRespLen))
	}
	var wf WebFinger
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
	return
}

func (svc service) FetchActor(ctx context.Context, addr vocab.IRI) (actor vocab.Actor, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, string(addr), nil)
	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", "application/activity+json")
		req.Header.Add("User-Agent", svc.hostSelf)
		resp, err = svc.clientHttp.Do(req)
	}
	var data []byte
	if err == nil {
		data, err = io.ReadAll(io.LimitReader(resp.Body, limitRespLen))
	}
	if err == nil {
		err = json.Unmarshal(data, &actor)
	}
	return
}

func (svc service) RequestFollow(ctx context.Context, host string, addr, inbox vocab.IRI) (err error) {
	//
	follow := vocab.Activity{
		Type:    vocab.FollowType,
		Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
		Actor:   vocab.IRI(fmt.Sprintf("https://%s/actor", svc.hostSelf)),
		Object:  addr,
	}
	var followData []byte
	followData, err = json.Marshal(follow)
	//
	var req *http.Request
	if err == nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, string(inbox), bytes.NewReader(followData))
	}
	req.Header.Set("Content-Type", "application/ld+json; profile=\"http://www.w3.org/ns/activitystreams\"")
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("User-Agent", svc.hostSelf)
	req.Header.Set("Host", host)
	now := time.Now().UTC()
	req.Header.Set("Date", now.Format(http.TimeFormat))
	//
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
		err = signer.SignRequest(privKey, fmt.Sprintf("https://%s/actor#main-key", svc.hostSelf), req, followData)
		if err != nil {
			err = fmt.Errorf("failed to sign the follow request: %w", err)
		}
	}
	//
	var resp *http.Response
	if err == nil {
		resp, err = svc.clientHttp.Do(req)
	}
	var respData []byte
	if err == nil {
		respData, err = io.ReadAll(io.LimitReader(resp.Body, limitRespLen))
		if err == nil && resp.StatusCode >= 300 {
			err = fmt.Errorf("follow response status: %d, headers: %+v, content:\n%s\n", resp.StatusCode, resp.Header, string(respData))
		}
	}
	//
	return
}
