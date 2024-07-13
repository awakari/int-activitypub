package main

import (
	"context"
	"fmt"
	"github.com/awakari/client-sdk-go/api"
	apiGrpc "github.com/awakari/int-activitypub/api/grpc"
	apiHttp "github.com/awakari/int-activitypub/api/http"
	"github.com/awakari/int-activitypub/api/http/handler"
	"github.com/awakari/int-activitypub/api/http/reader"
	"github.com/awakari/int-activitypub/config"
	"github.com/awakari/int-activitypub/model"
	"github.com/awakari/int-activitypub/service"
	"github.com/awakari/int-activitypub/service/activitypub"
	"github.com/awakari/int-activitypub/service/converter"
	"github.com/awakari/int-activitypub/service/writer"
	"github.com/awakari/int-activitypub/storage"
	"github.com/gin-gonic/gin"
	vocab "github.com/go-ap/activitypub"
	apiProm "github.com/prometheus/client_golang/api"
	apiPromV1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/writeas/go-nodeinfo"
	"log/slog"
	"net/http"
	"os"
)

func main() {

	// init config and logger
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

	// init storage
	var stor storage.Storage
	stor, err = storage.NewStorage(context.TODO(), cfg.Db)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the storage: %s", err))
	}
	stor = storage.NewLocalCache(stor, cfg.Db.Table.Following.Cache.Size, cfg.Db.Table.Following.Cache.Ttl)
	defer stor.Close()

	// awakari API client
	var clientAwk api.Client
	clientAwk, err = api.
		NewClientBuilder().
		WriterUri(cfg.Api.Writer.Uri).
		SubscriptionsUri(cfg.Api.Interests.Uri).
		Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize the Awakari API client: %s", err))
	}
	defer clientAwk.Close()
	log.Info("initialized the Awakari API client")

	// prometheus client
	clientProm, err := apiProm.NewClient(apiProm.Config{
		Address: cfg.Api.Prometheus.Uri,
	})
	var ap apiPromV1.API
	switch err {
	case nil:
		ap = apiPromV1.NewAPI(clientProm)
	default:
		log.Warn(fmt.Sprintf("Failed to connect Prometheus @ %s: %s", cfg.Api.Prometheus.Uri, err))
		err = nil
	}

	clientHttp := &http.Client{}
	svcActivityPub := activitypub.NewService(clientHttp, cfg.Api.Http.Host, []byte(cfg.Api.Key.Private), ap)
	svcActivityPub = activitypub.NewServiceLogging(svcActivityPub, log)

	svcConv := converter.NewService(cfg.Api.EventType, fmt.Sprintf("https://%s", cfg.Api.Http.Host), vocab.ActivityVocabularyType(cfg.Api.Actor.Type))
	svcConv = converter.NewLogging(svcConv, log)

	svcWriter := writer.NewService(clientAwk, cfg.Api.Writer.Backoff)
	svcWriter = writer.NewLogging(svcWriter, log)

	// init websub reader
	svcReader := reader.NewService(clientHttp, cfg.Api.Reader.Uri)
	svcReader = reader.NewServiceLogging(svcReader, log)
	urlCallbackBase := fmt.Sprintf(
		"%s://%s:%d%s",
		cfg.Api.Reader.CallBack.Protocol,
		cfg.Api.Reader.CallBack.Host,
		cfg.Api.Reader.CallBack.Port,
		cfg.Api.Reader.CallBack.Path,
	)

	svc := service.NewService(stor, svcActivityPub, cfg.Api.Http.Host, svcConv, svcWriter, svcReader, urlCallbackBase)
	svc = service.NewLogging(svc, log)

	log.Info(fmt.Sprintf("starting to listen the gRPC API @ port #%d...", cfg.Api.Port))
	go func() {
		if err = apiGrpc.Serve(cfg.Api.Port, svc); err != nil {
			panic(err)
		}
	}()

	// nodeinfo
	cfgNodeInfo := nodeinfo.Config{
		BaseURL: fmt.Sprintf("https://%s", cfg.Api.Http.Host),
		InfoURL: "/api/nodeinfo",
		Metadata: nodeinfo.Metadata{
			NodeName:        cfg.Api.Node.Name,
			NodeDescription: cfg.Api.Node.Description,
			Private:         false,
			Software: nodeinfo.SoftwareMeta{
				HomePage: "https://awakari.com",
				GitHub:   "https://github.com/awakari/int-activitypub",
				Follow:   "https://github.com/awakari/int-activitypub",
			},
		},
		Protocols: []nodeinfo.NodeProtocol{
			nodeinfo.ProtocolActivityPub,
		},
		Services: nodeinfo.Services{
			Inbound: []nodeinfo.NodeService{
				nodeinfo.ServiceAtom,
				nodeinfo.ServiceRSS,
				"telegram",
			},
			Outbound: []nodeinfo.NodeService{
				nodeinfo.ServiceRSS,
				"telegram",
			},
		},
		Software: nodeinfo.SoftwareInfo{
			Name:    "Awakari",
			Version: "1.0.0",
		},
	}
	nodeInfo := nodeinfo.NewService(cfgNodeInfo, svcActivityPub)

	// actor
	actor := vocab.Actor{
		ID:   vocab.ID(fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host)),
		Type: vocab.ActivityVocabularyType(cfg.Api.Actor.Type),
		Name: vocab.DefaultNaturalLanguageValue(cfg.Api.Actor.Name),
		Context: vocab.ItemCollection{
			vocab.IRI(model.NsAs),
			vocab.IRI("https://w3id.org/security/v1"),
		},
		Icon: vocab.Image{
			MediaType: "image/png",
			Type:      vocab.ImageType,
			URL:       vocab.IRI("https://awakari.com/logo-color-64.png"),
		},
		Summary: vocab.DefaultNaturalLanguageValue(
			"<p>Awakari is a free service that discovers and follows interesting Fediverse publishers on behalf of own users. " +
				"The service accepts public only messages and filters these to fulfill own user interest queries.</p>" +
				"<p>Before accepting any publisher's data, Awakari requests to follow them. " +
				"The acceptance means publisher's <i>explicit consent</i> to process their public messages, like most of other Fediverse servers do. " +
				"If you don't agree with the following, please don't accept the follow request or remove Awakari from your followers.</p>" +
				"Contact: <a href=\"mailto:awakari@awakari.com\">awakari@awakari.com</a><br/>" +
				"Donate: <a href=\"https://awakari.com/donation.html\">https://awakari.com/donation.html</a><br/>" +
				"Opt-Out: <a href=\"https://github.com/awakari/.github/blob/master/OPT-OUT.md\">https://github.com/awakari/.github/blob/master/OPT-OUT.md</a><br/>" +
				"Privacy: <a href=\"https://awakari.com/privacy.html\">https://awakari.com/privacy.html</a><br/>" +
				"Source: <a href=\"https://github.com/awakari/int-activitypub\">https://github.com/awakari/int-activitypub</a><br/>" +
				"Terms: <a href=\"https://awakari.com/tos.html\">https://awakari.com/tos.html</a></p>",
		),
		URL:               vocab.IRI("https://awakari.com/activitypub"),
		Inbox:             vocab.IRI(fmt.Sprintf("https://%s/inbox", cfg.Api.Http.Host)),
		Outbox:            vocab.IRI(fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host)),
		Following:         vocab.IRI(fmt.Sprintf("https://%s/following", cfg.Api.Http.Host)),
		Followers:         vocab.IRI(fmt.Sprintf("https://%s/followers", cfg.Api.Http.Host)),
		PreferredUsername: vocab.DefaultNaturalLanguageValue(cfg.Api.Actor.Name),
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
				ID:   vocab.ID("https://awakari.com/activitypub"),
				URL:  vocab.IRI("https://awakari.com/activitypub"),
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
	// TODO get the url prefix from config
	ha := handler.NewActorHandler(actor, actorExtraAttrs, clientAwk, "https://awakari.com/sub-details.html?id=", cfg.Api)

	// WebFinger
	wfDefault := apiHttp.WebFinger{
		Subject: fmt.Sprintf("acct:%s@%s", cfg.Api.Actor.Name, cfg.Api.Http.Host),
		Links: []apiHttp.WebFingerLink{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: fmt.Sprintf("https://%s/actor", cfg.Api.Http.Host),
			},
		},
	}
	hwf := handler.NewWebFingerHandler(wfDefault, cfg.Api.Http.Host, clientAwk)

	// handlers for inbox, outbox, following, followers
	hi := handler.NewInboxHandler(svcActivityPub, svc, cfg.Api.Http.Host)
	ho := handler.NewOutboxHandler(svcReader, svcConv, fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host))
	hoDummy := handler.NewDummyCollectionHandler(vocab.OrderedCollectionPage{
		ID:      vocab.IRI(fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host)),
		Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
		PartOf:  vocab.IRI(fmt.Sprintf("https://%s/outbox", cfg.Api.Http.Host)),
		First:   vocab.IRI(fmt.Sprintf("https://%s/outbox?page=1", cfg.Api.Http.Host)),
	})
	hFollowing := handler.NewFollowingHandler(stor, fmt.Sprintf("https://%s/following", cfg.Api.Http.Host))
	hFollowers := handler.NewFollowersHandler(svcReader, fmt.Sprintf("https://%s/followers", cfg.Api.Http.Host))

	r := gin.Default()
	r.GET("/.well-known/webfinger", hwf.Handle)
	r.GET("/actor/:id", ha.Handle)
	r.GET("/actor", ha.Handle)
	r.POST("/inbox/:id", hi.Handle)
	r.POST("/inbox", hi.Handle)
	r.GET("/outbox/:id", ho.Handle)
	r.GET("/outbox", hoDummy.Handle)
	r.GET("/following/:id", handler.NewDummyCollectionHandler(vocab.OrderedCollectionPage{
		ID:      vocab.IRI(fmt.Sprintf("https://%s/dummy/inbox", cfg.Api.Http.Host)),
		Context: vocab.IRI("https://www.w3.org/ns/activitystreams"),
		PartOf:  vocab.IRI(fmt.Sprintf("https://%s/dummy/inbox", cfg.Api.Http.Host)),
		First:   vocab.IRI(fmt.Sprintf("https://%s/dummy/inbox?page=1", cfg.Api.Http.Host)),
	}).Handle)
	r.GET("/following", hFollowing.Handle)
	r.GET("/followers/:id", hFollowers.Handle)
	r.GET(nodeinfo.NodeInfoPath, func(ctx *gin.Context) {
		nodeInfo.NodeInfoDiscover(ctx.Writer, ctx.Request)
		return
	})
	r.GET(cfgNodeInfo.InfoURL, func(ctx *gin.Context) {
		nodeInfo.NodeInfo(ctx.Writer, ctx.Request)
		return
	})
	r.GET("/", ha.Handle)
	log.Info(fmt.Sprintf("starting to listen the HTTP API @ port #%d...", cfg.Api.Http.Port))
	go func() {
		err = r.Run(fmt.Sprintf(":%d", cfg.Api.Http.Port))
		if err != nil {
			panic(err)
		}
	}()

	hc := handler.NewCallbackHandler(cfg.Api.Reader.Uri, cfg.Api.Http.Host, svcConv, svcActivityPub)

	log.Info(fmt.Sprintf("starting to listen the HTTP API @ port #%d...", cfg.Api.Reader.CallBack.Port))
	internalCallbacks := gin.Default()
	internalCallbacks.
		GET(cfg.Api.Reader.CallBack.Path, hc.Confirm).
		POST(cfg.Api.Reader.CallBack.Path, hc.Deliver)
	err = internalCallbacks.Run(fmt.Sprintf(":%d", cfg.Api.Reader.CallBack.Port))
	if err != nil {
		panic(err)
	}
}
