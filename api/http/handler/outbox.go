package handler

import (
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
)

type dummyCollHandler struct {
	coll map[string]any
}

func NewDummyCollectionHandler(coll vocab.OrderedCollectionPage) Handler {
	collFixed, _ := apiHttp.FixContext(coll)
	return dummyCollHandler{
		coll: collFixed,
	}
}

func (o dummyCollHandler) Handle(ctx *gin.Context) {
	ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
	ctx.JSON(http.StatusOK, o.coll)
	return
}
