package activitypub

import (
	"context"
	"fmt"
	"github.com/awakari/int-activitypub/util"
	vocab "github.com/go-ap/activitypub"
	"github.com/writeas/go-nodeinfo"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) ResolveActorLink(ctx context.Context, host, name string) (self vocab.IRI, err error) {
	switch name {
	case "fail":
		err = ErrActorWebFinger
	default:
		self = vocab.IRI(fmt.Sprintf("https://%s/users/%s", host, name))
	}
	return
}

func (m mock) FetchActor(ctx context.Context, self vocab.IRI, pubKeyId string) (a vocab.Actor, tags util.ObjectTags, err error) {
	switch self {
	case "https://fail.social/users/johndoe":
		err = ErrActorFetch
	case "https://privacy.social/users/nobot1":
		a.ID = self
		a.Name = vocab.DefaultNaturalLanguageValue("Bots Hater1")
		a.Inbox = vocab.IRI(fmt.Sprintf("%s/inbox", self))
		a.Summary = vocab.DefaultNaturalLanguageValue("Please #nobot otherwise I will complain")
	case "https://privacy.social/users/nobot2":
		a.ID = self
		a.Name = vocab.DefaultNaturalLanguageValue("Bots Hater2")
		a.Inbox = vocab.IRI(fmt.Sprintf("%s/inbox", self))
		tags = util.ObjectTags{
			Tag: []util.ActivityTag{
				{
					Name: "#nobot",
				},
			},
		}
	default:
		a.ID = self
		a.Name = vocab.DefaultNaturalLanguageValue("John Doe")
		a.Inbox = vocab.IRI(fmt.Sprintf("%s/inbox", self))
	}
	return
}

func (m mock) SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI, pubKeyId string) (err error) {
	switch inbox {
	case "https://host.fail/users/johndoe/inbox":
		err = ErrActivitySend
	}
	return
}

func (m mock) IsOpenRegistration() (bool, error) {
	return true, nil
}

func (m mock) Usage() (u nodeinfo.Usage, err error) {
	return
}
