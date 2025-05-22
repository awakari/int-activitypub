package reader

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/model"
	"github.com/bytedance/sonic"
	ceProto "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2/event"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Service interface {
	Subscribe(ctx context.Context, interestId, groupId, userId, url string, interval time.Duration) (err error)
	Subscription(ctx context.Context, interestId, groupId, userId, url string) (cb Subscription, err error)
	Unsubscribe(ctx context.Context, interestId, groupId, userId, url string) (err error)
	CountByInterest(ctx context.Context, interestId, groupId, userId string) (count int64, err error)
	Feed(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error)
}

type service struct {
	clientHttp    *http.Client
	uriBase       string
	tokenInternal string
}

const keyHubCallback = "hub.callback"
const KeyHubMode = "hub.mode"
const KeyHubTopic = "hub.topic"
const modeSubscribe = "subscribe"
const modeUnsubscribe = "unsubscribe"
const fmtTopicUri = "%s/v1/sub/%s/%s"
const FmtJson = "json"
const fmtReadUri = "%s/v1/sub/%s/%s?limit=%d"

var ErrInternal = errors.New("internal failure")
var ErrConflict = errors.New("conflict")
var ErrNotFound = errors.New("not found")

func NewService(clientHttp *http.Client, uriBase, tokenInternal string) Service {
	return service{
		clientHttp:    clientHttp,
		uriBase:       uriBase,
		tokenInternal: tokenInternal,
	}
}

func (svc service) Subscribe(ctx context.Context, interestId, groupId, userId, urlCallback string, interval time.Duration) (err error) {
	err = svc.updateCallback(ctx, interestId, groupId, userId, urlCallback, modeSubscribe, interval)
	return
}

func (svc service) Unsubscribe(ctx context.Context, interestId, groupId, userId, urlCallback string) (err error) {
	err = svc.updateCallback(ctx, interestId, groupId, userId, urlCallback, modeUnsubscribe, 0)
	return
}

func (svc service) updateCallback(ctx context.Context, interestId, groupId, userId, urlCallback, mode string, interval time.Duration) (err error) {

	topicUri := fmt.Sprintf(fmtTopicUri, svc.uriBase, FmtJson, interestId)
	data := url.Values{
		keyHubCallback: {
			urlCallback,
		},
		KeyHubMode: {
			mode,
		},
		KeyHubTopic: {
			topicUri,
		},
	}
	reqUri := fmt.Sprintf("%s/v2?format=%s&interestId=%s", svc.uriBase, FmtJson, interestId)
	if interval > 0 && mode == modeSubscribe {
		reqUri += "&interval=" + interval.String()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUri, strings.NewReader(data.Encode()))
	var resp *http.Response
	if err == nil {
		req.Header.Set("Authorization", "Bearer "+svc.tokenInternal)
		req.Header.Set(model.KeyGroupId, groupId)
		req.Header.Set(model.KeyUserId, userId)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err = svc.clientHttp.Do(req)
	}

	switch err {
	case nil:
		switch resp.StatusCode {
		case http.StatusAccepted, http.StatusNoContent:
		case http.StatusNotFound:
			err = fmt.Errorf("%w: callback not found for the subscription %s", ErrConflict, interestId)
		case http.StatusConflict:
			err = fmt.Errorf("%w: callback already registered for the subscription %s", ErrConflict, interestId)
		default:
			defer resp.Body.Close()
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 0x1000))
			err = fmt.Errorf("%w: unexpected create callback response %d, %s", ErrInternal, resp.StatusCode, string(body))
		}
	default:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}

func (svc service) Subscription(ctx context.Context, interestId, groupId, userId, urlCallback string) (cb Subscription, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/v2?interestId=%s&url=%s", svc.uriBase, interestId, base64.URLEncoding.EncodeToString([]byte(urlCallback))),
		http.NoBody,
	)
	var resp *http.Response
	if err == nil {
		req.Header.Set("Authorization", "Bearer "+svc.tokenInternal)
		req.Header.Set(model.KeyGroupId, groupId)
		req.Header.Set(model.KeyUserId, userId)
		resp, err = svc.clientHttp.Do(req)
	}
	switch err {
	case nil:
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&cb)
			if err != nil {
				err = fmt.Errorf("%w: %s", ErrInternal, err)
			}
		case http.StatusNotFound:
			err = ErrNotFound
		default:
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 0x1000))
			err = fmt.Errorf("%w: response %d, %s", ErrInternal, resp.StatusCode, string(body))
		}
	default:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}

func (svc service) CountByInterest(ctx context.Context, interestId, groupId, userId string) (count int64, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v2?interestId=%s", svc.uriBase, interestId), http.NoBody)
	var resp *http.Response
	if err == nil {
		req.Header.Set("Authorization", "Bearer "+svc.tokenInternal)
		req.Header.Set(model.KeyGroupId, groupId)
		req.Header.Set(model.KeyUserId, userId)
		resp, err = svc.clientHttp.Do(req)
	}
	switch err {
	case nil:
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&count)
			if err != nil {
				err = fmt.Errorf("%w: %s", ErrInternal, err)
			}
		case http.StatusNotFound:
			err = ErrNotFound
		default:
			err = fmt.Errorf("%w: response status %d", ErrInternal, resp.StatusCode)
		}
	default:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}

func (svc service) Feed(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error) {
	u := fmt.Sprintf(fmtReadUri, svc.uriBase, FmtJson, interestId, limit)
	var resp *http.Response
	resp, err = svc.clientHttp.Get(u)
	switch err {
	case nil:
		defer resp.Body.Close()
		var evts []*ce.Event
		err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&evts)
		if err != nil {
			err = fmt.Errorf("%w: failed to deserialize the request payload: %s", ErrInternal, err)
			return
		}
		for _, evt := range evts {
			var evtProto *pb.CloudEvent
			evtProto, err = ceProto.ToProto(evt)
			if err != nil {
				err = fmt.Errorf("%w: failed to deserialize the event %s: %s", ErrInternal, evt.ID(), err)
				break
			}
			last = append(last, evtProto)
		}
	default:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}
