package handler

import (
	"fmt"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
	"net/url"
)

type following struct {
	stor    storage.Storage
	baseUrl string
}

const keyCursor = "cursor"
const defaultLimit = 10
const defaultOrder = model.OrderAsc

var defaultFilter = model.Filter{}

func NewFollowingHandler(stor storage.Storage, baseUrl string) Handler {
	return following{
		stor:    stor,
		baseUrl: baseUrl,
	}
}

func (f following) Handle(ctx *gin.Context) {
	cursor := ctx.Query(keyCursor)
	if cursor != "" {
		var err error
		cursor, err = url.QueryUnescape(cursor)
		if err != nil {
			ctx.String(http.StatusBadRequest, fmt.Sprintf("invalid cursor value \"%s\", cause: %s", cursor, err))
			return
		}
	}
	count, err := f.stor.Count(ctx)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to estimate count of sources: %s", err))
		return
	}
	page, err := f.stor.List(ctx, defaultFilter, defaultLimit, cursor, defaultOrder)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to list sources: %s", err))
		return
	}
	ocp := vocab.OrderedCollectionPage{
		ID:         vocab.IRI(f.baseUrl + "?" + keyCursor + "=" + cursor),
		Type:       "OrderedCollectionPage",
		Context:    vocab.IRI("https://www.w3.org/ns/activitystreams"),
		PartOf:     vocab.IRI(f.baseUrl),
		TotalItems: uint(count),
		First:      vocab.IRI(f.baseUrl),
	}
	for _, src := range page {
		ocp.OrderedItems = append(ocp.OrderedItems, vocab.IRI(src))
	}
	if len(page) > 0 {
		next := page[len(page)-1]
		next = url.QueryEscape(next)
		ocp.Next = vocab.IRI(f.baseUrl + "?" + keyCursor + "=" + next)
	}
	ocpFixed, cs := apiHttp.FixContext(ocp)
	ctx.Writer.Header().Set("content-type", apiHttp.ContentTypeActivity)
	ctx.Writer.Header().Set("etag", fmt.Sprintf("W/\"%x\"", cs))
	ctx.JSON(http.StatusOK, ocpFixed)
	return
}
