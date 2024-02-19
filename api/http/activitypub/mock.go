package activitypub

import (
	"context"
	"fmt"
	vocab "github.com/go-ap/activitypub"
)

type mock struct {
}

func NewServiceMock() Service {
	return mock{}
}

func (m mock) FetchActor(ctx context.Context, self vocab.IRI) (a vocab.Actor, err error) {
	switch self {
	case "https://fail.social/users/johndoe":
		err = ErrActorFetch
	default:
		a.ID = self
		a.Name = vocab.DefaultNaturalLanguageValue("John Doe")
		a.Inbox = vocab.IRI(fmt.Sprintf("%s/inbox", self))
	}
	return
}

func (m mock) SendActivity(ctx context.Context, a vocab.Activity, inbox vocab.IRI) (err error) {
	switch inbox {
	case "https://host.fail/users/johndoe/inbox":
		err = ErrActivitySend
	}
	return
}
