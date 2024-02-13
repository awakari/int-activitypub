package handler

import (
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/gin-gonic/gin"
	"net/http"
)

type webFingerHandler struct {
	wf apiHttp.WebFinger
}

func NewWebFingerHandler(wf apiHttp.WebFinger) Handler {
	return webFingerHandler{
		wf: wf,
	}
}

func (w webFingerHandler) Handle(ctx *gin.Context) {
	ctx.Writer.Header().Add("Content-Type", "application/jrd+json; charset=utf-8")
	ctx.JSON(http.StatusOK, w.wf)
	return
}
