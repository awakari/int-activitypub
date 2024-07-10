package handler

import (
	"fmt"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/api/http/reader"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
)

type outboxHandler struct {
	svcReader reader.Service
	svcConv   converter.Service
	baseUrl   string
}

const maxPageLen = 100

func NewOutboxHandler(svcReader reader.Service, svcConv converter.Service, baseUrl string) Handler {
	return outboxHandler{
		svcReader: svcReader,
		svcConv:   svcConv,
		baseUrl:   baseUrl,
	}
}

func (oh outboxHandler) Handle(ctx *gin.Context) {

	id := ctx.Param("id")
	evts, err := oh.svcReader.Read(ctx, id, maxPageLen)
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	u := oh.baseUrl + "/" + id
	uPage := u + "?page=true"
	resultColl := vocab.OrderedCollectionNew(vocab.ID(u))
	var result vocab.ActivityObject
	page := ctx.Query("page")
	switch page {
	case "":
		result = resultColl
		resultColl.Context = vocab.IRI(model.NsAs)
		resultColl.TotalItems = uint(len(evts))
		resultColl.First = vocab.IRI(uPage)
		resultColl.Last = resultColl.First
	default:
		resultCollPage := vocab.OrderedCollectionPageNew(resultColl)
		result = resultCollPage
		resultCollPage.Context = vocab.IRI(model.NsAs)
		resultCollPage.TotalItems = uint(len(evts))
		resultCollPage.ID = vocab.ID(uPage)
		resultCollPage.Prev = vocab.IRI(uPage)
		resultCollPage.Next = resultCollPage.Prev
		for _, evt := range evts {
			a, err := oh.svcConv.ConvertEventToActivity(ctx, evt, id, nil)
			switch err {
			case nil:
				resultCollPage.OrderedItems = append(resultCollPage.OrderedItems, a)
			default:
				fmt.Printf("failed to convert event %s to activity, skipping: %s\n", evt.Id, err)
			}
		}
	}

	d, cs := apiHttp.FixContext(result)
	ctx.Writer.Header().Set("content-type", apiHttp.ContentTypeActivity)
	ctx.Writer.Header().Set("etag", fmt.Sprintf("W/\"%x\"", cs))
	ctx.JSON(http.StatusOK, d)
	return
}
