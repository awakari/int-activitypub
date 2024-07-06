package handler

import (
	"errors"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	"github.com/awakari/client-sdk-go/api/grpc/subscriptions"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/model"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"google.golang.org/grpc/metadata"
	"net/http"
)

type actorHandler struct {
	actorDefault             vocab.Actor
	actorDefaultFixed        map[string]any
	extraAttrs               map[string]any
	clientAwk                api.Client
	urlPrefixInterestDetails string
	cfgApi                   config.ApiConfig
}

func NewActorHandler(
	actorDefault vocab.Actor,
	extraAttrs map[string]any,
	clientAwk api.Client,
	urlPrefixInterestDetails string,
	cfgApi config.ApiConfig,
) (h Handler) {
	aFixed := apiHttp.FixContext(actorDefault)
	for k, v := range extraAttrs {
		aFixed[k] = v
	}
	h = actorHandler{
		actorDefault:             actorDefault,
		actorDefaultFixed:        aFixed,
		extraAttrs:               extraAttrs,
		clientAwk:                clientAwk,
		urlPrefixInterestDetails: urlPrefixInterestDetails,
		cfgApi:                   cfgApi,
	}
	return
}

func (ah actorHandler) Handle(ctx *gin.Context) {
	accept := ctx.Request.Header.Get("Accept")
	id := ctx.Param("id")
	switch id {
	case "":
		ah.handleDefault(ctx, accept)
	default:
		ah.handleInterest(ctx, accept, id)
	}
	return
}

func (ah actorHandler) handleDefault(ctx *gin.Context, accept string) {
	switch accept {
	case "text/html", "application/xhtml+xml", "text/xml", "application/xml":
		ctx.Writer.Header().Add("Content-Type", "text/html; charset=utf-8")
		ctx.String(http.StatusOK, `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Awakari</title>
    <meta charset="utf-8">
</head>
<body>
	<h1>Awakari</h1>
	<p>
		Awakari is a free service that discovers and follows interesting Fediverse publishers on behalf of own users.
		The service accepts and filters public only messages to fulfill own user interest queries.</p>
	<p>
		Before accepting any publisher's data, Awakari requests to follow them.
		The acceptance means publisher's consent to process their public messages, like most of other Fediverse servers do.
		If you don't agree with the following, please don't accept the follow request or remove Awakari from your followers.</p>
	<p>
		<a href="mailto:awakari@awakari.com">Contact</a><br/>
		<a href="https://awakari.com/donation.html">Donate</a><br/>
		<a href="https://awakari.com/login.html">Login</a><br/>
		<a href="https://github.com/awakari/.github/blob/master/OPT-OUT.md">Opt-Out</a><br/>
		<a href="https://awakari.com/privacy.html">Privacy</a><br/>
		<a href="https://github.com/awakari/int-activitypub">Source</a><br/>
		<a href="https://awakari.com/tos.html">Terms</a>
	</p>
</body>`)
	default:
		ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
		ctx.JSON(http.StatusOK, ah.actorDefaultFixed)
	}
	return
}

func (ah actorHandler) handleInterest(ctx *gin.Context, accept, id string) {
	switch accept {
	case "text/html", "application/xhtml+xml", "text/xml", "application/xml":
		ctx.Redirect(http.StatusMovedPermanently, ah.urlPrefixInterestDetails+id)
	default:
		ctxAwk := metadata.AppendToOutgoingContext(ctx, model.KeyGroupId, model.GroupIdDefault)
		d, err := ah.clientAwk.ReadSubscription(ctxAwk, model.UserIdDefault, id)
		switch {
		case err == nil:
			ctx.Writer.Header().Add("Content-Type", "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"")
			actor := ah.actorDefault // derive the default actor
			actor.ID = vocab.ID(fmt.Sprintf("https://%s/actor/%s", ah.cfgApi.Http.Host, id))
			actor.Name = vocab.DefaultNaturalLanguageValue(id)
			actor.Summary = vocab.DefaultNaturalLanguageValue(fmt.Sprintf("Awakari Interest: %s", d.Description))
			actor.URL = vocab.IRI(ah.urlPrefixInterestDetails + id)
			actor.Inbox = vocab.IRI(fmt.Sprintf("https://%s/inbox/%s", ah.cfgApi.Http.Host, id))
			actor.Outbox = vocab.IRI(fmt.Sprintf("https://%s/outbox/%s", ah.cfgApi.Http.Host, id))
			actor.Following = vocab.IRI(fmt.Sprintf("https://%s/following/%s", ah.cfgApi.Http.Host, id))
			actor.Followers = vocab.IRI(fmt.Sprintf("https://%s/followers/%s", ah.cfgApi.Http.Host, id))
			actor.PreferredUsername = vocab.DefaultNaturalLanguageValue(id)
			actor.Endpoints = &vocab.Endpoints{
				SharedInbox: vocab.IRI(fmt.Sprintf("https://%s/inbox/%s", ah.cfgApi.Http.Host, id)),
			}
			actor.Attachment = vocab.ItemCollection{
				vocab.Page{
					Name: vocab.DefaultNaturalLanguageValue("homepage"),
					ID:   vocab.ID(ah.urlPrefixInterestDetails + id),
					URL:  vocab.IRI(ah.urlPrefixInterestDetails + id),
				},
			}
			aFixed := apiHttp.FixContext(actor)
			for k, v := range ah.extraAttrs {
				aFixed[k] = v
			}
			ctx.JSON(http.StatusOK, aFixed)
		case errors.Is(err, subscriptions.ErrNotFound):
			ctx.String(http.StatusNotFound, "public interest does not exist: %s", id)
		default:
			ctx.String(http.StatusInternalServerError, err.Error())
		}
	}
	return
}
