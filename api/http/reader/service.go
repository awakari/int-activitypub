package reader

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	ceProto "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2/event"
	"io"
	"net/http"
)

type Service interface {
	CreateCallback(ctx context.Context, subId, url string) (err error)
	GetCallback(ctx context.Context, subId, url string) (cb Callback, err error)
	DeleteCallback(ctx context.Context, subId, url string) (err error)
	CountByInterest(ctx context.Context, interestId string) (count int64, err error)
	Read(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error)
}

type service struct {
	clientHttp *http.Client
	uriBase    string
}

const keyHubCallback = "hub.callback"
const KeyHubMode = "hub.mode"
const KeyHubTopic = "hub.topic"
const modeSubscribe = "subscribe"
const modeUnsubscribe = "unsubscribe"
const fmtTopicUri = "%s/sub/%s/%s"
const FmtJson = "json"
const fmtReadUri = "%s/sub/%s/%s?limit=%d"

var ErrInternal = errors.New("internal failure")
var ErrConflict = errors.New("conflict")
var ErrNotFound = errors.New("not found")

func NewService(clientHttp *http.Client, uriBase string) Service {
	return service{
		clientHttp: clientHttp,
		uriBase:    uriBase,
	}
}

func (svc service) CreateCallback(ctx context.Context, subId, callbackUrl string) (err error) {
	err = svc.updateCallback(ctx, subId, callbackUrl, modeSubscribe)
	return
}

func (svc service) CountByInterest(ctx context.Context, interestId string) (count int64, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/callbacks/list/%s", svc.uriBase, interestId), http.NoBody)
	var resp *http.Response
	if err == nil {
		resp, err = svc.clientHttp.Do(req)
	}
	switch err {
	case nil:
		defer resp.Body.Close()
		switch resp.StatusCode {
		case http.StatusOK:
			var cbl CallbackList
			err = sonic.ConfigDefault.NewDecoder(resp.Body).Decode(&cbl)
			switch err {
			case nil:
				count = cbl.Count
			default:
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

func (svc service) GetCallback(ctx context.Context, subId, url string) (cb Callback, err error) {
	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/callbacks/%s/%s", svc.uriBase, subId, base64.URLEncoding.EncodeToString([]byte(url))), http.NoBody)
	var resp *http.Response
	if err == nil {
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
			err = fmt.Errorf("%w: response status %d", ErrInternal, resp.StatusCode)
		}
	default:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}

func (svc service) DeleteCallback(ctx context.Context, subId, callbackUrl string) (err error) {
	err = svc.updateCallback(ctx, subId, callbackUrl, modeUnsubscribe)
	return
}

func (svc service) updateCallback(_ context.Context, subId, url, mode string) (err error) {
	topicUri := fmt.Sprintf(fmtTopicUri, svc.uriBase, FmtJson, subId)
	data := map[string][]string{
		keyHubCallback: {
			url,
		},
		KeyHubMode: {
			mode,
		},
		KeyHubTopic: {
			topicUri,
		},
	}
	var resp *http.Response
	resp, err = svc.clientHttp.PostForm(topicUri, data)
	switch err {
	case nil:
		switch resp.StatusCode {
		case http.StatusAccepted, http.StatusNoContent:
		case http.StatusNotFound:
			err = fmt.Errorf("%w: callback not found for the subscription %s", ErrConflict, subId)
		case http.StatusConflict:
			err = fmt.Errorf("%w: callback already registered for the subscription %s", ErrConflict, subId)
		default:
			defer resp.Body.Close()
			respBody, _ := io.ReadAll(resp.Body)
			err = fmt.Errorf("%w: unexpected create callback response %d, %s", ErrInternal, resp.StatusCode, string(respBody))
		}
	default:
		err = fmt.Errorf("%w: %s", ErrInternal, err)
	}
	return
}

func (svc service) Read(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error) {
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
