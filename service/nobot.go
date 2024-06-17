package service

import (
	"github.com/awakari/int-activitypub/util"
)

const NoBot = "#nobot"

func ActorHasNoBotTag(o util.ObjectTags) (contains bool) {
	for _, t := range o.Tag {
		if t.Name == NoBot {
			contains = true
			break
		}
	}
	return
}

func ActivityHasNoBotTag(a util.ActivityTags) (contains bool) {
	for _, t := range a.Tag {
		if t.Name == NoBot {
			contains = true
			break
		}
	}
	for _, t := range a.Object.Tag {
		if t.Name == NoBot {
			contains = true
			break
		}
	}
	return
}
