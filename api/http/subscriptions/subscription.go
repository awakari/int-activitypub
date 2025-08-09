package subscriptions

type Subscription struct {
	Url    string `json:"url"`
	Format string `json:"fmt"`
}

const QueryParamFollower = "follower"
