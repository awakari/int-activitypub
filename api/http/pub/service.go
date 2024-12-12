package pub

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/model"
	"github.com/bytedance/sonic"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
)

type Service interface {
	Publish(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error)
}

type service struct {
	clientHttp *http.Client
	url        string
	token      string
}

type payloadResp struct {
	AckCount uint32 `json:"ackCount"`
}

const valContentTypeJson = "application/json"

var ErrNoAck = errors.New("publishing is not acknowledged")
var ErrNoAuth = errors.New("unauthenticated request")
var ErrInvalid = errors.New("invalid request")
var ErrLimitReached = errors.New("publishing limit reached")

func NewService(clientHttp *http.Client, url, token string) Service {
	return service{
		clientHttp: clientHttp,
		url:        url,
		token:      token,
	}
}

func (svc service) Publish(ctx context.Context, evt *pb.CloudEvent, groupId, userId string) (err error) {

	var reqData []byte
	reqData, err = protojson.Marshal(evt)

	var req *http.Request
	if err == nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, svc.url, bytes.NewReader(reqData))
	}

	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", valContentTypeJson)
		req.Header.Add("Authorization", "Bearer "+svc.token)
		req.Header.Add("Content-Type", valContentTypeJson)
		req.Header.Add(model.KeyGroupId, groupId)
		req.Header.Add(model.KeyUserId, userId)
		resp, err = svc.clientHttp.Do(req)
	}

	if err == nil {
		switch resp.StatusCode {
		case http.StatusServiceUnavailable:
			err = fmt.Errorf("%w: %s", ErrNoAck, evt.Id)
		case http.StatusUnauthorized:
			err = ErrNoAuth
		case http.StatusRequestTimeout:
			err = fmt.Errorf("%w: %s", ErrNoAck, evt.Id)
		case http.StatusBadRequest:
			err = fmt.Errorf("%w: %s", ErrInvalid, evt.Id)
		case http.StatusTooManyRequests:
			err = fmt.Errorf("%w: %s", ErrLimitReached, evt.Id)
		}
	}

	var respData []byte
	if err == nil {
		defer resp.Body.Close()
		respData, err = io.ReadAll(resp.Body)
	}

	var p payloadResp
	if err == nil {
		err = sonic.Unmarshal(respData, &p)
	}

	if err == nil && p.AckCount < 1 {
		err = fmt.Errorf("%w: %s", ErrNoAck, evt.Id)
	}

	return
}
