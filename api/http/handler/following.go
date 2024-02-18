package handler

import (
	"fmt"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
	"net/url"
)

type following struct {
	stor storage.Storage
}

const keyCursor = "cursor"
const defaultLimit = 10
const defaultOrder = model.OrderAsc

var defaultFilter = model.Filter{}

func NewFollowingHandler(stor storage.Storage) Handler {
	return following{
		stor: stor,
	}
}

func (f following) Handle(ctx *gin.Context) {
	baseUrl := fmt.Sprintf("%s://%s%s", ctx.Request.URL.Scheme, ctx.Request.URL.Host, ctx.Request.URL.Path)
	cursor := ctx.Query(keyCursor)
	if cursor != "" {
		var err error
		cursor, err = url.QueryUnescape(cursor)
		if err != nil {
			ctx.String(http.StatusInternalServerError, fmt.Sprintf("invalid cursor value \"%s\", cause: %s", cursor, err))
			return
		}
	}
	page, err := f.stor.List(ctx, defaultFilter, defaultLimit, cursor, defaultOrder)
	switch err {
	case nil:
		ocp := vocab.OrderedCollectionPage{
			ID:      vocab.ID(ctx.Request.URL.String()),
			Type:    "OrderedCollectionPage",
			Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
			PartOf:  vocab.IRI(baseUrl),
		}
		for _, src := range page {
			ocp.OrderedItems = append(ocp.OrderedItems, vocab.IRI(src))
		}
		if len(page) > 0 {
			next := page[len(page)-1]
			next = url.QueryEscape(next)
			ocp.Next = vocab.IRI(fmt.Sprintf("%s?%s=%s", baseUrl, keyCursor, next))
		}
		ctx.JSON(http.StatusOK, ocp)
	default:
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to list sources: %s", err))
	}
	return
}
