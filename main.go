package main

import (
	"context"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/int-activitypub/api/grpc"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/api/http/handler"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/service"
	activitypub2 "github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/awakari/int-activitypub/service/mastodon"
	"github.com/awakari/int-activitypub/service/writer"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
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
	stor, err = storage.NewStorage(context.TODO(), cfg.Db)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the storage: %s", err))
	}
	stor = storage.NewLocalCache(stor, cfg.Db.Table.Following.Cache.Size, cfg.Db.Table.Following.Cache.Ttl)
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
	svcActivityPub := activitypub2.NewService(clientHttp, cfg.Api.Http.Host, []byte(cfg.Api.Key.Private))
	svcActivityPub = activitypub2.NewServiceLogging(svcActivityPub, log)
	//
	svcConv := converter.NewService()
	svcConv = converter.NewLogging(svcConv, log)
	//
	svcWriter := writer.NewService(clientAwk, cfg.Api.Writer.Backoff)
	svcWriter = writer.NewLogging(svcWriter, log)
	//
	svc := service.NewService(stor, svcActivityPub, cfg.Api.Http.Host, svcConv, svcWriter)
	svc = service.NewLogging(svc, log)
	//
	svcMstdn := mastodon.NewService(clientHttp, cfg.Api.Http.Host, cfg.Search.Mastodon, svc, svcWriter)
	svcMstdn = mastodon.NewServiceLogging(svcMstdn, log)
	go func() {
		for {
			_ = svcMstdn.ConsumeLiveStreamPublic(context.Background())
		}
	}()
	//
	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	go func() {
		if err = apiGrpc.Serve(cfg.Api.Port, svc, svcMstdn); err != nil {
			panic(err)
		}
	}()
	//
	actor := vocab.Actor{
		ID:   vocab.ID(fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host)),
		Type: vocab.ServiceType,
		Name: vocab.DefaultNaturalLanguageValue("AwakariBot"),
		Context: vocab.ItemCollection{
			vocab.IRI("https://www.w3.org/ns/activitystreams"),
			vocab.IRI("https://w3id.org/security/v1"),
		},
		Icon: vocab.Image{
			MediaType: "image/png",
			Type:      vocab.ImageType,
			URL:       vocab.IRI("https://awakari.com/logo-color-256.png"),
		},
		Image: vocab.Image{
			MediaType: "image/png",
			Type:      vocab.ImageType,
			URL:       vocab.IRI("https://awakari.com/logo-color-1024.png"),
		},
		Summary:           vocab.DefaultNaturalLanguageValue("Awakari ActivityPub Bot"),
		URL:               vocab.IRI("https://awakari.com"),
		Inbox:             vocab.IRI(fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host)),
		Outbox:            vocab.IRI(fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host)),
		Following:         vocab.IRI(fmt.Sprintf("https://%s/following", cfg.Api.Http.Host)),
		Followers:         vocab.IRI(fmt.Sprintf("https://%s/followers", cfg.Api.Http.Host)),
		PreferredUsername: vocab.DefaultNaturalLanguageValue("AwakariBot"),
		Endpoints: &vocab.Endpoints{
			SharedInbox: vocab.IRI(fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host)),
		},
		PublicKey: vocab.PublicKey{
			ID:           vocab.ID(fmt.Sprintf("https://%s/actor#main-key", cfg.Api.Http.Host)),
			Owner:        vocab.IRI(fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host)),
			PublicKeyPem: cfg.Api.Key.Public,
		},
		Attachment: vocab.ItemCollection{
			vocab.Page{
				Name: vocab.DefaultNaturalLanguageValue("Home"),
				ID:   vocab.ID("https://awakari.com"),
				URL:  vocab.IRI("https://awakari.com"),
			},
			vocab.Page{
				Name: vocab.DefaultNaturalLanguageValue("GitHub"),
				ID:   vocab.ID("https://github.com/awakari"),
				URL:  vocab.IRI("https://githun.com/awakari"),
			},
			vocab.Page{
				Name: vocab.DefaultNaturalLanguageValue("Telegram Bot"),
				ID:   vocab.ID("https://t.me/AwakariBot"),
				URL:  vocab.IRI("https://t.me/AwakariBot"),
			},
		},
	}
	actorExtraAttrs := map[string]any{
		"manuallyApprovesFollowers": false,
		"discoverable":              true,
		"indexable":                 true,
		"memorial":                  false,
	}
	ha := handler.NewActorHandler(actor, actorExtraAttrs)
	wf := apiHttp.WebFinger{
		Subject: fmt.Sprintf("acct:AwakariBot@%s", cfg.Api.Http.Host),
		Links: []apiHttp.WebFingerLink{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host),
			},
		},
	}
	hwf := handler.NewWebFingerHandler(wf)
	hi := handler.NewInboxHandler(svcActivityPub, svc)
	outboxItems := vocab.ItemCollection{
		vocab.Article{
			ID:        "https://awakari.com/articles/why-not-rss.html",
			URL:       vocab.IRI("https://awakari.com/articles/why-not-rss.html"),
			Name:      vocab.DefaultNaturalLanguageValue("Why Not RSS?"),
			Summary:   vocab.DefaultNaturalLanguageValue("Today many of us want to get updates timely from the services they use: news, shops, jobs, etc. But visiting every service every minute is wasting of the time. One can increase the update period and do it daily. Unfortunately, this means the important update is not early enough. Even worse, the important update may be gone already and lost."),
			Published: time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
		},
		vocab.Article{
			ID:        "https://awakari.com/articles/spacetime-search.html",
			URL:       vocab.IRI("https://awakari.com/articles/spacetime-search.html"),
			Name:      vocab.DefaultNaturalLanguageValue("On Search In Space And Tie"),
			Summary:   vocab.DefaultNaturalLanguageValue("The role of a chance in our life is great. It's always about finding a relevant opportunity. Whether it's a dream job, car, home or a partner."),
			Published: time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC),
		},
		vocab.Article{
			ID:        "https://awakari.com/articles/beyond-rss.html",
			URL:       vocab.IRI("https://awakari.com/articles/beyond-rss.html"),
			Name:      vocab.DefaultNaturalLanguageValue("Beyond RSS"),
			Summary:   vocab.DefaultNaturalLanguageValue("The idea of the Awakari service is to filter important events from unlimited number of various sources. This article is about the ways it uses to extract a structured data from the Internet beyond RSS feeds and Telegram channels."),
			Published: time.Date(2024, 2, 9, 0, 0, 0, 0, time.UTC),
		},
		vocab.Article{
			ID:        "https://awakari.com/articles/awakari-goes-social/index.html",
			URL:       vocab.IRI("https://awakari.com/articles/awakari-goes-social/index.html"),
			Name:      vocab.DefaultNaturalLanguageValue("Beyond RSS"),
			Summary:   vocab.DefaultNaturalLanguageValue("Today more and more services support ActivityPub to exchange activities in the decentralized world also known as Fediverse. The applications are not limited to social networks and blogs. There are also image, music, video sharing services and more.\n\n   By treating these activities as source events Awakari brings the whole new dimension to the Fediverse world. With Awakari, one can track any activity matching own criteria from unlimited number of publishers and services in addition to the existing source types like RSS and Telegram."),
			Published: time.Date(2024, 2, 22, 0, 0, 0, 0, time.UTC),
		},
	}
	outboxItemsToPublish := vocab.ItemCollection{
		outboxItems[0],
		outboxItems[1],
		outboxItems[2],
		outboxItems[3],
	}
	for _, p := range outboxItemsToPublish {
		a := vocab.Activity{
			ID:     vocab.IRI(fmt.Sprintf("%s#create", p.GetID())),
			Type:   vocab.CreateType,
			URL:    vocab.IRI(fmt.Sprintf("%s#create", p.GetID())),
			Actor:  vocab.IRI(fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host)),
			Object: p,
		}
		err = svcActivityPub.SendActivity(context.TODO(), a, "https://mastodon.social/inbox")
		switch err {
		case nil:
			log.Info(fmt.Sprintf("Published the outbox activity %+v", a))
		default:
			log.Info(fmt.Sprintf("Failed to publish the outbox activity %+v: %s", a, err))
		}
	}
	ho := handler.NewOutboxHandler(vocab.OrderedCollectionPage{
		ID:           vocab.IRI(fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host)),
		Context:      vocab.IRI("https://www.w3.org/ns/activitystreams"),
		OrderedItems: outboxItems,
		TotalItems:   uint(len(outboxItems)),
	})
	hFollowing := handler.NewFollowingHandler(stor)
	//
	r := gin.Default()
	r.GET("/.well-known/webfinger", hwf.Handle)
	r.GET("/actor", ha.Handle)
	r.POST("/inbox", hi.Handle)
	r.GET("/outbox", ho.Handle)
	r.GET("/following", hFollowing.Handle)
	log.Info(fmt.Sprintf("starting to listen the HTTP API @ port #%d...", cfg.Api.Http.Port))
	err = r.Run(fmt.Sprintf(":%d", cfg.Api.Http.Port))
	if err != nil {
		panic(err)
	}
}
