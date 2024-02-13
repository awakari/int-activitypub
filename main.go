package main

import (
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/int-activitypub/api/grpc"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	"log/slog"
	"net/http"
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
	clientHttp := &http.Client{}
	svcActivityPub := apiHttp.NewService(clientHttp, cfg.Api.Http.Host, []byte(cfg.Api.Key.Private))
	svcActivityPub = apiHttp.NewServiceLogging(svcActivityPub, log)
	//
	svc := service.NewService(stor, svcActivityPub)
	svc = service.NewLogging(svc, log)
	//
	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	go func() {
		if err = apiGrpc.Serve(cfg.Api.Port, svc); err != nil {
			panic(err)
		}
	}()
	//
	a := vocab.Actor{
		ID:   vocab.ID(fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host)),
		Type: vocab.ServiceType,
		Name: vocab.DefaultNaturalLanguageValue("awakari"),
		Context: vocab.IRIs{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		Icon: vocab.Image{
			MediaType: "image/png",
			Type:      vocab.ImageType,
			URL:       vocab.IRI("https://awakari.com/logo-color-64.png"),
		},
		Image: vocab.Image{
			MediaType: "image/svg+xml",
			Type:      vocab.ImageType,
			URL:       vocab.IRI("https://awakari.com/logo-color.svg"),
		},
		Summary:           vocab.DefaultNaturalLanguageValue("Awakari ActivityPub Bot"),
		URL:               vocab.IRI("https://awakari.com"),
		Inbox:             vocab.IRI(fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host)),
		Outbox:            vocab.IRI(fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host)),
		Following:         vocab.IRI(fmt.Sprintf("https://%s/following", cfg.Api.Http.Host)),
		Followers:         vocab.IRI(fmt.Sprintf("https://%s/followers", cfg.Api.Http.Host)),
		PreferredUsername: vocab.DefaultNaturalLanguageValue("awakari"),
		Endpoints: &vocab.Endpoints{
			SharedInbox: vocab.IRI(fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host)),
		},
		PublicKey: vocab.PublicKey{
			ID:           vocab.ID(fmt.Sprintf("https://%s/actor#main-key", cfg.Api.Http.Host)),
			Owner:        vocab.IRI(fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host)),
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
	hi := apiHttp.NewInboxHandler(svcActivityPub)
	//
	r := gin.Default()
	r.GET("/actor", ha.Handle)
	r.GET("/.well-known/webfinger", hwf.Handle)
	r.POST("/inbox", hi.Handle)
	log.Info(fmt.Sprintf("starting to listen the HTTP API @ port #%d...", cfg.Api.Http.Port))
	err = r.Run(fmt.Sprintf(":%d", cfg.Api.Http.Port))
	if err != nil {
		panic(err)
	}
}
