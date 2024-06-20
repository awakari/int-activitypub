package handler

import (
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"net/http"
)

type actorHandler struct {
	a map[string]any
}

func NewActorHandler(a vocab.Actor, extraAttrs map[string]any) (h Handler) {
	aFixed := apiHttp.FixContext(a)
	for k, v := range extraAttrs {
		aFixed[k] = v
	}
	h = actorHandler{
		a: aFixed,
	}
	return
}

func (ah actorHandler) Handle(ctx *gin.Context) {
	accept := ctx.Request.Header.Get("Accept")
	switch accept {
	case "text/html", "application/xhtml+xml", "text/xml", "application/xml":
		ctx.Writer.Header().Add("Content-Type", "text/html; charset=utf-8")
		ctx.String(http.StatusOK, `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Awakari ActivityPub</title>
    <meta charset="utf-8">
</head>
<body>
	<h1>Awakari</h1>
	<h2>ActivityPub</h2>
	<p>
		Awakari is a free service that discovers and follows interesting Fediverse publishers on behalf of own users.
		The service accepts public only messages and filters these to fulfill own user interest queries.</p>
	<p>
		Before accepting any publisher's data, Awakari requests to follow them.
		The acceptance means publisher's <i>explicit consent</i> to process their public messages, like most of other Fediverse servers do.
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
		ctx.JSON(http.StatusOK, ah.a)
	}
	return
}
