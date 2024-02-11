package main

import (
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/int-activitypub/api/grpc"
	"github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	"log/slog"
	"os"
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
	a := http.Actor{
		Context: []string{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		Type:              "Person",
		Id:                fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host),
		Name:              "Awakari",
		PreferredUsername: "Awakari",
		Inbox:             fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host),
		Outbox:            fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host),
		Following:         fmt.Sprintf("https://%s/following", cfg.Api.Http.Host),
		Followers:         fmt.Sprintf("https://%s/followers", cfg.Api.Http.Host),
		PublicKey: http.ActorPublicKey{
			Id:           fmt.Sprintf("https://%s/actor#main-key", cfg.Api.Http.Host),
			Owner:        fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host),
			PublicKeyPem: cfg.Api.Key.Public,
		},
	}
	ha := http.NewActorHandler(a)
	//
	wf := http.WebFinger{
		Subject: fmt.Sprintf("acct:awakari@%s", cfg.Api.Http.Host),
		Links: []http.WebFingerLink{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: "https://mastodon.social/users/awakari",
			},
		},
	}
	hwf := http.NewWebFingerHandler(wf)
	//
	r := gin.Default()
	r.GET("/actor", ha.Handle)
	r.GET("/.well-known/webfinger", hwf.Handle)
	log.Info(fmt.Sprintf("starting to listen the HTTP API @ port #%d...", cfg.Api.Http.Port))
	err = r.Run(fmt.Sprintf(":%d", cfg.Api.Http.Port))
	if err != nil {
		panic(err)
	}
}
