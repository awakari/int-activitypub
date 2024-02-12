package main

import (
	"bytes"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/int-activitypub/api/grpc"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	"github.com/superseriousbusiness/httpsig"
	"golang.org/x/crypto/ssh"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	//
	cfg, err := config.NewConfigFromEnv()
	if err != nil {
		panic(fmt.Sprintf("failed to load the config from env: %s", err))
	}
	//
	opts := slog.HandlerOptions{
		Level: slog.Level(cfg.Log.Level),
	}
	log := slog.New(slog.NewTextHandler(os.Stdout, &opts))
	log.Info("starting the update for the feeds")
	//
	var stor storage.Storage
	//stor, err = storage.NewStorage(context.TODO(), cfg.Db)
	stor = storage.NewStorageMock()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the storage: %s", err))
	}
	defer stor.Close()
	//
	var clientAwk api.Client
	clientAwk, err = api.
		NewClientBuilder().
		WriterUri(cfg.Api.Writer.Uri).
		Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the Awakari API client: %s", err))
	}
	defer clientAwk.Close()
	log.Info("initialized the Awakari API client")
	//
	svc := service.NewService(stor)
	svc = service.NewLogging(svc, log)
	//
	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	go func() {
		if err = apiGrpc.Serve(cfg.Api.Port, svc); err != nil {
			panic(err)
		}
	}()
	//
	a := apiHttp.Actor{
		Context: []string{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		Type:              "Service",
		Id:                fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host),
		Name:              "awakari",
		PreferredUsername: "awakari",
		Inbox:             fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host),
		Outbox:            fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host),
		Following:         fmt.Sprintf("https://%s/following", cfg.Api.Http.Host),
		Followers:         fmt.Sprintf("https://%s/followers", cfg.Api.Http.Host),
		Endpoints: apiHttp.ActorEndpoints{
			SharedInbox: fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host),
		},
		Url:     "https://awakari.com",
		Summary: "Awakari ActivityPub Bot",
		Icon: apiHttp.ActorMedia{
			MediaType: "image/png",
			Type:      "Image",
			Url:       "https://awakari.com/logo-color-64.png",
		},
		Image: apiHttp.ActorMedia{
			MediaType: "image/svg+xml",
			Type:      "Image",
			Url:       "https://awakari.com/logo-color.svg",
		},
		PublicKey: apiHttp.ActorPublicKey{
			Id:           fmt.Sprintf("https://%s/actor#main-key", cfg.Api.Http.Host),
			Owner:        fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host),
			PublicKeyPem: cfg.Api.Key.Public,
		},
	}
	ha := apiHttp.NewActorHandler(a)
	//
	wf := apiHttp.WebFinger{
		Subject: fmt.Sprintf("acct:awakari@%s", cfg.Api.Http.Host),
		Links: []apiHttp.WebFingerLink{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host),
			},
		},
	}
	hwf := apiHttp.NewWebFingerHandler(wf)
	//
	r := gin.Default()
	r.GET("/actor", ha.Handle)
	r.GET("/.well-known/webfinger", hwf.Handle)
	r.POST("/inbox", func(ctx *gin.Context) {
		data, err := io.ReadAll(io.LimitReader(ctx.Request.Body, 65536))
		if err != nil {
			panic(err)
		}
		fmt.Printf("Incoming Activity:\nHeaders: %+v\nBody: %s\n", ctx.Request.Header, string(data))
		ctx.Status(http.StatusOK)
	})
	log.Info(fmt.Sprintf("starting to listen the HTTP API @ port #%d...", cfg.Api.Http.Port))
	go func() {
		time.Sleep(5 * time.Second)
		err = follow([]byte(cfg.Api.Key.Private), cfg.Api.Http.Host)
		if err != nil {
			fmt.Printf("failed to follow: %s\n", err)
		}
	}()
	err = r.Run(fmt.Sprintf(":%d", cfg.Api.Http.Port))
	if err != nil {
		panic(err)
	}
}

func follow(privKey []byte, host string) (err error) {
	data := []byte(fmt.Sprintf(`{
        "@context": "https://www.w3.org/ns/activitystreams",
        "type": "Follow",
        "actor": "https://%s/actor",
        "object": "https://mastodon.social/users/akurilov"
    }`, host))
	req, err := http.NewRequest("POST", "https://mastodon.social/users/akurilov/inbox", bytes.NewReader(data))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/ld+json; profile=\"http://www.w3.org/ns/activitystreams\"")
	req.Header.Add("Accept-Charset", "utf-8")
	req.Header.Add("User-Agent", host)
	req.Header.Set("Host", "mastodon.social")
	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	digestAlgorithm := httpsig.DigestSha256
	headersToSign := []string{httpsig.RequestTarget, "host", "date", "digest"}
	signer, _, err := httpsig.NewSigner(prefs, digestAlgorithm, headersToSign, httpsig.Signature, 120)
	if err != nil {
		return
	}
	now := time.Now().UTC()
	req.Header.Set("Date", now.Format(http.TimeFormat))
	priv, err := ssh.ParseRawPrivateKey(privKey)
	if err != nil {
		fmt.Printf("failed to parse the private key: %s\n", err)
		return
	}
	err = signer.SignRequest(priv, fmt.Sprintf("https://%s/actor#main-key", host), req, data)
	if err != nil {
		fmt.Printf("failed to sign the follow request: %s\n", err)
		return
	}

	client := &http.Client{}
	fmt.Printf("Follow, headers: %+v\n", req.Header)
	resp, err := client.Do(req)
	if err == nil {
		respData, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Printf("Follow, response status: %d, headers: %+v, content:\n%s\n", resp.StatusCode, resp.Header, string(respData))
		}
	}
	return
}
