package mastodon

type Results struct {
	Statuses []Status
}

type Status struct {
	Sensitive bool `json:"sensitive"`
	Account   Account
}

type Account struct {
	Acct string `json:"acct"`
	Uri  string `json:"uri"`
}
