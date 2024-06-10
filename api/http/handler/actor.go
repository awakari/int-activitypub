package handler

import (
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
)

type actorHandler struct {
	a map[string]any
}

func NewActorHandler(a vocab.Actor, extraAttrs map[string]any) (h Handler) {
	aFixed := apiHttp.FixContext(a)
	for k, v := range extraAttrs {
		aFixed[k] = v
	}
	h = actorHandler{
		a: aFixed,
	}
	return
}

func (ah actorHandler) Handle(ctx *gin.Context) {
	accept := ctx.Request.Header.Get("Accept")
	switch accept {
	case "application/json", "application/ld+json":
		ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		ctx.JSON(http.StatusOK, ah.a)
	default:
		ctx.Redirect(http.StatusMovedPermanently, "https://awakari.com/login.html")
	}
	return
}
