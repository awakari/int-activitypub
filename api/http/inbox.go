package http

import (
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"github.com/superseriousbusiness/httpsig"
	"io"
	"net/http"
)

type inboxHandler struct {
	svc Service
}

func NewInboxHandler(svc Service) Handler {
	return inboxHandler{
		svc: svc,
	}
}

func (h inboxHandler) Handle(ctx *gin.Context) {
	//
	activity, actor, err := h.verify(ctx)
	if err != nil {
		fmt.Printf("Inbox request verification failed: %s\n", err)
		ctx.String(http.StatusBadRequest, err.Error())
		return
	}
	//
	fmt.Printf("Inbox activity %+v from actor: %+v\n", activity.Type, actor.URL.GetLink())
	ctx.Status(http.StatusOK)
	//
	return
}

func (h inboxHandler) verify(ctx *gin.Context) (activity vocab.Activity, actor vocab.Actor, err error) {
	//
	req := ctx.Request
	//
	var data []byte
	if err == nil {
		data, err = io.ReadAll(io.LimitReader(req.Body, limitRespLen))
	}
	if err == nil {
		err = json.Unmarshal(data, &activity)
	}
	if err == nil {
		actor, err = h.svc.FetchActor(ctx, activity.Actor.GetLink())
	}
	//
	var verifier httpsig.Verifier
	if err == nil {
		verifier, err = httpsig.NewVerifier(req)
	}
	if err == nil {
		pubKeyId := verifier.KeyId()
		if pubKeyId != actor.PublicKey.ID.String() {
			err = fmt.Errorf("the actor pub key id %s doesn't match the request's one %s", actor.PublicKey.ID, pubKeyId)
		}
	}
	//
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
