package reader

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	ceProto "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2/event"
	"net/http"
)

type Service interface {
	Feed(ctx context.Context, interestId string, limit int) (last []*pb.CloudEvent, err error)
}

type service struct {
	clientHttp *http.Client
	uriBase    string
}

const FmtJson = "json"
const fmtReadUri = "%s/v1/sub/%s/%s?limit=%d"

var ErrInternal = errors.New("internal failure")

func NewService(clientHttp *http.Client, uriBase string) Service {
	return service{
		clientHttp: clientHttp,
		uriBase:    uriBase,
	}
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
