package handler

import (
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
)

type outboxHandler struct {
	coll map[string]any
}

func NewOutboxHandler(coll vocab.OrderedCollectionPage) Handler {
	return outboxHandler{
		coll: apiHttp.FixContext(coll),
	}
}

func (o outboxHandler) Handle(ctx *gin.Context) {
	ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	ctx.JSON(http.StatusOK, o.coll)
	return
}
