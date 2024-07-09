package reader

type Callback struct {
	Url    string `json:"url"`
	Format string `json:"fmt"`
}

type CallbackList struct {
	Count int64 `json:"count"`
}

const QueryParamFollower = "follower"
