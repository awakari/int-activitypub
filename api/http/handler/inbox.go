package handler

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"github.com/superseriousbusiness/httpsig"
	"io"
	"net/http"
)

type inboxHandler struct {
	svcActivityPub activitypub.Service
	svc            service.Service
}

const limitReqBodyLen = 262_144

func NewInboxHandler(svcActivityPub activitypub.Service, svc service.Service) Handler {
	return inboxHandler{
		svcActivityPub: svcActivityPub,
		svc:            svc,
	}
}

func (h inboxHandler) Handle(ctx *gin.Context) {

	req := ctx.Request
	data, err := io.ReadAll(io.LimitReader(req.Body, limitReqBodyLen))
	if err != nil {
		fmt.Printf("Inbox request read failure: %s\n", err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	var activity vocab.Activity
	err = json.Unmarshal(data, &activity)
	if err != nil {
		fmt.Printf("Inbox request unmarshal failure: %s\n", err)
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	t := activity.Type
	if t == "" || t == vocab.DeleteType && activity.Actor.GetID() == activity.Object.GetID() {
		ctx.Status(http.StatusOK)
		return
	}

	var actor vocab.Actor
	actor, err = h.svcActivityPub.FetchActor(ctx, activity.Actor.GetLink())
	if err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	err = h.verify(ctx, data, actor)
	if err != nil {
		fmt.Printf("Inbox request verification failed: %s\n", err)
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}

	err = h.svc.HandleActivity(ctx, actor, activity)
	switch {
	case errors.Is(err, storage.ErrNotFound), errors.Is(err, service.ErrInvalid):
		ctx.String(http.StatusBadRequest, err.Error())
		return
	case err != nil:
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Status(http.StatusOK)
	return
}

func (h inboxHandler) verify(ctx *gin.Context, data []byte, actor vocab.Actor) (err error) {
	var verifier httpsig.Verifier
	if err == nil {
		verifier, err = httpsig.NewVerifier(ctx.Request)
	}
	if err == nil {
		pubKeyId := verifier.KeyId()
		if pubKeyId != actor.PublicKey.ID.String() {
			err = fmt.Errorf("the actor pub key %+v doesn't match the request's one %s, activity: %s", actor.PublicKey, pubKeyId, string(data))
		}
	}
	var pubKeyDer *pem.Block
	if err == nil {
		pubKeyDer, _ = pem.Decode([]byte(actor.PublicKey.PublicKeyPem))
		if pubKeyDer == nil {
			err = fmt.Errorf("failed to decode author public key PEM: %s", actor.PublicKey.PublicKeyPem)
		}
	}
	var pubKey crypto.PublicKey
	if err == nil {
		pubKey, err = x509.ParsePKIXPublicKey(pubKeyDer.Bytes)
	}
	// The verifier will verify the Digest in addition to the HTTP signature
	if err == nil {
		err = verifier.Verify(pubKey, httpsig.RSA_SHA256)
	}
	return
}
