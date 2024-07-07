package handler

import (
	"fmt"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/api/http/reader"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
)

type followers struct {
	svcReader reader.Service
	baseUrl   string
}

func NewFollowersHandler(svcReader reader.Service, baseUrl string) Handler {
	return followers{
		svcReader: svcReader,
		baseUrl:   baseUrl,
	}
}

func (hf followers) Handle(ctx *gin.Context) {
	interestId := ctx.Param("id")
	count, err := hf.svcReader.CountByInterest(ctx, interestId)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to estimate count of followers: %s", err))
		return
	}
	url := hf.baseUrl + "/" + interestId
	ocp := vocab.OrderedCollectionPage{
		ID:         vocab.IRI(url),
		Type:       "OrderedCollectionPage",
		Context:    vocab.IRI("https://www.w3.org/ns/activitystreams"),
		PartOf:     vocab.IRI(url),
		TotalItems: uint(count),
		First:      vocab.IRI(url + "?page=1"),
	}
	ocpFixed := apiHttp.FixContext(ocp)
	ctx.JSON(http.StatusOK, ocpFixed)
	return
}
