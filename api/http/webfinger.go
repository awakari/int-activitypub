package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type WebFinger struct {
	Subject string          `json:"subject"`
	Links   []WebFingerLink `json:"links"`
}

type WebFingerLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

type webFingerHandler struct {
	wf WebFinger
}

func NewWebFingerHandler(wf WebFinger) Handler {
	return webFingerHandler{
		wf: wf,
	}
}

func (w webFingerHandler) Handle(ctx *gin.Context) {
	fmt.Printf("WebFinger request headers: %+v\n", ctx.Request.Header)
	ctx.Writer.Header().Add("Content-Type", "application/jrd+json; charset=utf-8")
	ctx.JSON(http.StatusOK, w.wf)
	return
}
