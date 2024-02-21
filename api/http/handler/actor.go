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

func NewActorHandler(a vocab.Actor) (h Handler) {
	h = actorHandler{
		a: apiHttp.FixContext(a),
	}
	return
}

func (ah actorHandler) Handle(ctx *gin.Context) {
	ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	ctx.JSON(http.StatusOK, ah.a)
	return
}
