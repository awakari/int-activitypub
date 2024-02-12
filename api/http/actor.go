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
	Outbox            string         `json:"outbox"`
	Following         string         `json:"following"`
	Followers         string         `json:"followers"`
	Endpoints         ActorEndpoints `json:"endpoints"`
	Url               string         `json:"url"`
	Summary           string         `json:"summary"`
	Icon              ActorMedia     `json:"icon"`
	Image             ActorMedia     `json:"image"`
	PublicKey         ActorPublicKey `json:"publicKey"`
}

type ActorPublicKey struct {
	Id           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

type ActorEndpoints struct {
	SharedInbox string `json:"sharedInbox"`
}

type ActorMedia struct {
	MediaType string `json:"mediaType"`
	Type      string `json:"type"`
	Url       string `json:"url"`
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
