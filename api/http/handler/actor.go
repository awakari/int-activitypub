package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
	"strings"
)

type actorHandler struct {
	a vocab.Actor
}

func NewActorHandler(a vocab.Actor) (h Handler) {
	h = actorHandler{
		a: a,
	}
	return
}

func (ah actorHandler) Handle(ctx *gin.Context) {
	ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	ctx.String(http.StatusOK, marshalJsonAndFixContext(ah.a))
	return
}

func marshalJsonAndFixContext(a vocab.Actor) (txt string) {
	data, _ := json.Marshal(a)
	txt = string(data)
	txt = strings.Replace(txt, "\"context\":", "\"@context\":", -1)
	return
}
