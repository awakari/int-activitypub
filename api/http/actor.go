package http

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Actor struct {
	Context           []string       `json:"@context"`
	Type              string         `json:"type"`
	Id                string         `json:"id"`
	Name              string         `json:"name"`
	PreferredUsername string         `json:"preferredUsername"`
	Inbox             string         `json:"inbox"`
	PublicKey         ActorPublicKey `json:"publicKey"`
}

type ActorPublicKey struct {
	Id           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

type actorHandler struct {
	a Actor
}

func NewActorHandler(a Actor) (h Handler) {
	h = actorHandler{
		a: a,
	}
	return
}

func (ah actorHandler) Handle(ctx *gin.Context) {
	ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	ctx.JSON(http.StatusOK, ah.a)
	return
}
