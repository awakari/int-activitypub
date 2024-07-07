package reader

import (
	"net/url"
)

type Callback struct {
	Url    string `json:"url"`
	Format string `json:"fmt"`
}

type CallbackList struct {
	Count int64 `json:"count"`
}

const QueryParamFollower = "follower"

func MakeCallbackUrl(urlBase, follower string) string {
	return urlBase + "?" + QueryParamFollower + "=" + url.QueryEscape(follower)
}
