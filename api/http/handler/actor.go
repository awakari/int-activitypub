package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
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
	//
	ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	data, _ := json.Marshal(ah.a)
	// dirty fix the invalid "context" attribute, replace with "@context"
	raw := map[string]any{}
	_ = json.Unmarshal(data, &raw)
	v := raw["context"]
	delete(raw, "context")
	raw["@context"] = v
	//
	ctx.JSON(http.StatusOK, raw)
	return
}
