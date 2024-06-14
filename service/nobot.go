package service

import (
	"github.com/awakari/int-activitypub/util"
	vocab "github.com/go-ap/activitypub"
)

const NoBot = "#nobot"

func ActorHasNoBotTag(actor vocab.Actor) (contains bool) {
	for _, t := range actor.Tag {
		if t.IsObject() && t.(vocab.Object).Name.String() == NoBot {
			contains = true
			break
		}
	}
	return
}

func ActivityHasNoBotTag(tags util.ActivityTags) (contains bool) {
	for _, t := range tags.Tag {
		if t.Name == NoBot {
			contains = true
			break
		}
	}
	return
}
