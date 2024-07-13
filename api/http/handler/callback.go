package handler

import (
	"encoding/json"
	"fmt"
	"github.com/awakari/int-activitypub/api/http/reader"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/bytedance/sonic/utf8"
	ceProto "github.com/cloudevents/sdk-go/binding/format/protobuf/v2"
	"github.com/cloudevents/sdk-go/binding/format/protobuf/v2/pb"
	ce "github.com/cloudevents/sdk-go/v2/event"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type CallbackHandler interface {
	Confirm(ctx *gin.Context)
	Deliver(ctx *gin.Context)
}

type callbackHandler struct {
	topicPrefixBase string
	host            string
	svcConv         converter.Service
	svcAp           activitypub.Service
}

const keyHubChallenge = "hub.challenge"
const keyHubTopic = "hub.topic"
const linkSelfSuffix = ">; rel=\"self\""
const keyAckCount = "X-Ack-Count"

func NewCallbackHandler(topicPrefixBase, host string, svcConv converter.Service, svcAp activitypub.Service) CallbackHandler {
	return callbackHandler{
		topicPrefixBase: topicPrefixBase,
		host:            host,
		svcConv:         svcConv,
		svcAp:           svcAp,
	}
}

func (ch callbackHandler) Confirm(ctx *gin.Context) {
	topic := ctx.Query(keyHubTopic)
	challenge := ctx.Query(keyHubChallenge)
	if strings.HasPrefix(topic, ch.topicPrefixBase+"/sub/"+reader.FmtJson) {
		ctx.String(http.StatusOK, challenge)
	} else {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid topic: %s", topic))
	}
	return
}

func (ch callbackHandler) Deliver(ctx *gin.Context) {

	var topic string
	for k, vals := range ctx.Request.Header {
		if strings.ToLower(k) == "link" {
			for _, l := range vals {
				if strings.HasSuffix(l, linkSelfSuffix) && len(l) > len(linkSelfSuffix) {
					topic = l[1 : len(l)-len(linkSelfSuffix)]
				}
			}
		}
	}
	if topic == "" {
		ctx.String(http.StatusBadRequest, "self link header missing in the request")
		return
	}

	var interestId string
	topicParts := strings.Split(topic, "/")
	topicPartsLen := len(topicParts)
	if topicPartsLen > 0 {
		interestId = topicParts[topicPartsLen-1]
	}
	if interestId == "" {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid self link header value in the request: %s", topic))
		return
	}
	pubKeyId := fmt.Sprintf("https://%s/actor/%s#main-key", ch.host, interestId)

	followerUrl, err := url.QueryUnescape(ctx.Query(reader.QueryParamFollower))
	if err != nil || followerUrl == "" {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("follower parameter is missing or invalid: val=%s, err=%s", ctx.Query(reader.QueryParamFollower), err))
		return
	}

	var follower vocab.Actor
	follower, _, err = ch.svcAp.FetchActor(ctx, vocab.IRI(followerUrl), pubKeyId)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to resolve the follower %s: %s", follower, err))
		return
	}

	defer ctx.Request.Body.Close()
	var evts []*ce.Event
	err = json.NewDecoder(ctx.Request.Body).Decode(&evts)
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("failed to deserialize the request payload: %s", err))
		return
	}

	var countDelivered uint64
	for _, evt := range evts {
		var evtProto *pb.CloudEvent
		evtProto, err = ceProto.ToProto(evt)
		var dataTxt string
		if err == nil {
			err = evt.DataAs(&dataTxt)
		}
		if err == nil && utf8.ValidateString(dataTxt) {
			evtProto.Data = &pb.CloudEvent_TextData{
				TextData: dataTxt,
			}
		}
		var a vocab.Activity
		if err == nil {
			a, err = ch.svcConv.ConvertEventToActivity(ctx, evtProto, interestId, &follower)
		}
		if err == nil {
			err = ch.svcAp.SendActivity(ctx, a, follower.Inbox.GetLink(), pubKeyId)
		}
		if err != nil {
			break
		}
		countDelivered++
	}

	ctx.Writer.Header().Add(keyAckCount, strconv.FormatUint(countDelivered, 10))
	switch {
	case countDelivered < 1 && err != nil:
		ctx.String(http.StatusInternalServerError, err.Error())
	default:
		ctx.Status(http.StatusOK)
	}

	return
}
