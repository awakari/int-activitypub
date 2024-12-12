package interests

import (
	"context"
	"errors"
	"fmt"
	apiGrpc "github.com/awakari/int-activitypub/api/grpc/interests"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/model/interest"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
)

type Service interface {
	Read(ctx context.Context, groupId, userId, subId string) (subData interest.Data, err error)
}

type service struct {
	clientHttp *http.Client
	url        string
	token      string
}

var ErrNoAuth = errors.New("unauthenticated request")
var ErrNotFound = errors.New("interest not found")

func NewService(clientHttp *http.Client, url, token string) Service {
	return service{
		clientHttp: clientHttp,
		url:        url,
		token:      token,
	}
}

func (svc service) Read(ctx context.Context, groupId, userId, subId string) (subData interest.Data, err error) {

	var req *http.Request
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, svc.url+"/"+subId, nil)

	var resp *http.Response
	if err == nil {
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Authorization", "Bearer "+svc.token)
		req.Header.Add(model.KeyGroupId, groupId)
		req.Header.Add(model.KeyUserId, userId)
		resp, err = svc.clientHttp.Do(req)
	}

	if err == nil {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			err = ErrNoAuth
		case http.StatusNotFound:
			err = fmt.Errorf("%w: %s", ErrNotFound, subId)
		}
	}

	var respData []byte
	if err == nil {
		defer resp.Body.Close()
		respData, err = io.ReadAll(resp.Body)
	}

	var respProto apiGrpc.ReadResponse
	if err == nil {
		err = protojson.Unmarshal(respData, &respProto)
	}

	if err == nil {
		subData.Description = respProto.Description
		subData.Enabled = respProto.Enabled
		subData.Public = respProto.Public
		subData.Followers = respProto.Followers
		if respProto.Expires != nil && respProto.Expires.IsValid() {
			subData.Expires = respProto.Expires.AsTime()
		}
		if respProto.Created != nil && respProto.Created.IsValid() {
			subData.Created = respProto.Created.AsTime()
		}
		if respProto.Updated != nil && respProto.Updated.IsValid() {
			subData.Updated = respProto.Updated.AsTime()
		}
	}

	return
}
