package handler

import (
	"encoding/json"
	"fmt"
	"github.com/awakari/int-activitypub/api/http/reader"
	ce "github.com/cloudevents/sdk-go/v2/event"
	"github.com/gin-gonic/gin"
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
}

const keyHubChallenge = "hub.challenge"
const keyHubTopic = "hub.topic"
const linkSelfSuffix = ">; rel=\"self\""
const keyAckCount = "X-Ack-Count"

func NewCallbackHandler(topicPrefixBase string) CallbackHandler {
	return callbackHandler{
		topicPrefixBase: topicPrefixBase,
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

	follower, err := url.QueryUnescape(ctx.Query(reader.QueryParamFollower))
	if err != nil || follower == "" {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("follower parameter is missing or invalid: val=%s, err=%s", ctx.Query(reader.QueryParamFollower), err))
	}

	defer ctx.Request.Body.Close()
	var evts []*ce.Event
	err = json.NewDecoder(ctx.Request.Body).Decode(&evts)
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("failed to deserialize the request payload: %s", err))
		return
	}

	fmt.Printf("Deliver %d events to %s following the interest %s\n", len(evts), follower, interestId)
	ctx.Writer.Header().Add(keyAckCount, strconv.FormatUint(uint64(len(evts)), 10))
	ctx.Status(http.StatusOK)

	return
}
