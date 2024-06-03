package mastodon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/awakari/int-activitypub/service/writer"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	"github.com/google/uuid"
	"github.com/r3labs/sse/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error)
	ConsumeLiveStreamPublic() (err error)
}

type mastodon struct {
	clientHttp *http.Client
	userAgent  string
	cfg        config.MastodonConfig
	svc        service.Service
	w          writer.Service
}

const limitRespBodyLen = 1_048_576
const minFollowersCount = 10
const minPostCount = 10
const typeCloudEvent = "com.awakari.mastodon.v1"
const groupIdDefault = "default"
const streamSubDurationDefault = 1 * time.Hour

func NewService(clientHttp *http.Client, userAgent string, cfgMastodon config.MastodonConfig, svc service.Service, w writer.Service) Service {
	return mastodon{
		clientHttp: clientHttp,
		userAgent:  userAgent,
		cfg:        cfgMastodon,
		svc:        svc,
		w:          w,
	}
}

func (m mastodon) SearchAndAdd(ctx context.Context, subId, groupId, q string, limit uint32) (n uint32, err error) {
	var offset int
	for n < limit {
		reqQuery := "?q=" + url.QueryEscape(q) + "&type=statuses&offset=" + strconv.Itoa(offset) + "&limit=" + strconv.Itoa(int(limit))
		var req *http.Request
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, m.cfg.Endpoint.Search+reqQuery, nil)
		var resp *http.Response
		if err == nil {
			req.Header.Add("Accept", "application/json")
			req.Header.Add("Authorization", "Bearer "+m.cfg.Client.Token)
			req.Header.Add("User-Agent", m.userAgent)
			resp, err = m.clientHttp.Do(req)
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
			countResults := len(results.Statuses)
			if countResults == 0 {
				break
			}
			offset += countResults
			for _, s := range results.Statuses {
				var errReqFollow error
				if s.Sensitive {
					errReqFollow = fmt.Errorf("found account %s skip due to sensitive flag", s.Account.Uri)
				}
				if errReqFollow == nil && s.Account.FollowersCount < minFollowersCount {
					errReqFollow = fmt.Errorf("found account %s skip due low followers count %d", s.Account.Uri, s.Account.FollowersCount)
				}
				if errReqFollow == nil && s.Account.StatusesCount < minPostCount {
					errReqFollow = fmt.Errorf("found account %s skip due low post count %d", s.Account.Uri, s.Account.StatusesCount)
				}
				if errReqFollow == nil {
					_, errReqFollow = m.svc.RequestFollow(ctx, s.Account.Uri, groupId, "", subId, q)
				}
				if errReqFollow == nil {
					n++
				}
				if n > limit {
					break
				}
				err = errors.Join(err, errReqFollow)
			}
		}
	}
	return
}

func (m mastodon) ConsumeLiveStreamPublic() (err error) {
	client := sse.NewClient(m.cfg.Endpoint.Stream)
	client.Headers["Authorization"] = "Bearer " + m.cfg.Client.Token
	client.Headers["User-Agent"] = m.userAgent
	ctx, cancel := context.WithTimeout(context.Background(), streamSubDurationDefault)
	defer cancel()
	chSsEvts := make(chan *sse.Event)
	err = client.SubscribeChanWithContext(ctx, "", chSsEvts)
	if err == nil {
		defer client.Unsubscribe(chSsEvts)
		for {
			select {
			case ssEvt := <-chSsEvts:
				m.consumeLiveStreamEvent(ssEvt)
			case <-ctx.Done():
				err = ctx.Err()
			}
			if errors.Is(err, context.DeadlineExceeded) {
				err = nil
				break
			}
			if err != nil {
				break
			}
		}
	}
	return
}

func (m mastodon) consumeLiveStreamEvent(ssEvt *sse.Event) {
	if "update" == string(ssEvt.Event) {
		var st Status
		err := json.Unmarshal(ssEvt.Data, &st)
		if err != nil {
			fmt.Printf("failed to unmarshal the live stream event data: %s\nerror: %s\n", string(ssEvt.Data), err)
		}
		if st.Sensitive {
			return
		}
		if st.Visibility != "public" {
			return
		}
		evtAwk := m.convertStatus(st)
		userId := st.Account.Uri
		if userId == "" {
			userId = st.Account.Url
		}
		err = m.w.Write(context.TODO(), evtAwk, groupIdDefault, userId)
		if err != nil {
			fmt.Printf("failed to submit the live stream event, sse id=%s, awk id=%s, err=%s\n", string(ssEvt.ID), evtAwk.Id, err)
		}
	}
	return
}

func (m mastodon) convertStatus(st Status) (evtAwk *pb.CloudEvent) {
	evtAwk = &pb.CloudEvent{
		Id:          uuid.NewString(),
		Source:      m.cfg.Endpoint.Stream,
		SpecVersion: converter.CeSpecVersion,
		Type:        typeCloudEvent,
		Attributes: map[string]*pb.CloudEventAttributeValue{
			converter.CeKeySubject: {
				Attr: &pb.CloudEventAttributeValue_CeString{
					CeString: st.Account.DisplayName,
				},
			},
			converter.CeKeyTime: {
				Attr: &pb.CloudEventAttributeValue_CeTimestamp{
					CeTimestamp: timestamppb.New(st.CreatedAt.UTC()),
				},
			},
		},
		Data: &pb.CloudEvent_TextData{
			TextData: st.Content,
		},
	}
	if st.Language != "" {
		evtAwk.Attributes["language"] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: st.Language,
			},
		}
	}
	if st.Url != "" {
		evtAwk.Attributes[converter.CeKeyObjectUrl] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: st.Url,
			},
		}
	}
	var cats []string
	for _, t := range st.Tags {
		if t.Name != "" {
			cats = append(cats, t.Name)
		}
	}
	if len(cats) > 0 {
		evtAwk.Attributes[converter.CeKeyCategories] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: strings.Join(cats, " "),
			},
		}
	}
	if len(st.MediaAttachments) > 0 {
		att := st.MediaAttachments[0]
		evtAwk.Attributes[converter.CeKeyAttachmentType] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeString{
				CeString: att.Type,
			},
		}
		u := att.PreviewUrl
		if u == "" {
			u = att.Url
		}
		evtAwk.Attributes[converter.CeKeyAttachmentUrl] = &pb.CloudEventAttributeValue{
			Attr: &pb.CloudEventAttributeValue_CeUri{
				CeUri: u,
			},
		}
	}
	return
}
