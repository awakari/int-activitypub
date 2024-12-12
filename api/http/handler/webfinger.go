package handler

import (
	"errors"
	"fmt"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/api/http/interests"
	"github.com/awakari/int-activitypub/model"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type webFingerHandler struct {
	wfDefault    apiHttp.WebFinger
	hostSelf     string
	svcInterests interests.Service
}

func NewWebFingerHandler(wf apiHttp.WebFinger, hostSelf string, svcInterests interests.Service) Handler {
	return webFingerHandler{
		wfDefault:    wf,
		hostSelf:     hostSelf,
		svcInterests: svcInterests,
	}
}

func (w webFingerHandler) Handle(ctx *gin.Context) {
	r := ctx.Query(model.WebFingerKeyResource)
	switch r {
	case w.wfDefault.Subject:
		respond(ctx, w.wfDefault)
	default:
		w.handleNonDefault(ctx, r)
	}
	return
}

func (w webFingerHandler) handleNonDefault(ctx *gin.Context, r string) {
	rParts := strings.SplitN(r, model.AcctSep, 2)
	if len(rParts) != 2 {
		ctx.String(http.StatusBadRequest, "unrecognized resource format: %s", r)
		return
	}
	if rParts[1] != w.hostSelf {
		ctx.String(http.StatusBadRequest, "unrecognized resource host: %s", rParts[1])
		return
	}
	if !strings.HasPrefix(rParts[0], model.WebFingerPrefixAcct) {
		ctx.String(http.StatusBadRequest, "unrecognized resource account handle: %s", rParts[0])
		return
	}
	interestId := rParts[0][len(model.WebFingerPrefixAcct):]
	_, err := w.svcInterests.Read(ctx, model.GroupIdDefault, model.UserIdDefault, interestId)
	switch {
	case err == nil:
		wf := apiHttp.WebFinger{
			Subject: r,
			Links: []apiHttp.WebFingerLink{
				{
					Rel:  "self",
					Type: "application/activity+json",
					Href: fmt.Sprintf("https://%s/actor/%s", w.hostSelf, interestId),
				},
			},
		}
		respond(ctx, wf)
	case errors.Is(err, interests.ErrNotFound):
		ctx.String(http.StatusBadRequest, "interest doesn't exist: %s", interestId)
	default:
		ctx.String(http.StatusInternalServerError, err.Error())
	}
	return
}

func respond(ctx *gin.Context, wf apiHttp.WebFinger) {
	ctx.Writer.Header().Add("Content-Type", "application/jrd+json; charset=utf-8")
	ctx.JSON(http.StatusOK, wf)
	return
}
