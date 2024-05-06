package mastodon

type Results struct {
	Statuses []Status
}

type Status struct {
	Sensitive bool `json:"sensitive"`
	Account   Account
}

type Account struct {
	Acct           string `json:"acct"`
	DisplayName    string `json:"display_name"`
	Note           string `json:"note"`
	Uri            string `json:"uri"`
	FollowersCount int    `json:"followers_count"`
	StatusesCount  int    `json:"statuses_count"`
}