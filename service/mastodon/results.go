package mastodon

import "time"

type Results struct {
	Statuses []Status
}

type Status struct {
	CreatedAt        time.Time         `json:"created_at"`
	Visibility       string            `json:"visibility"`
	Language         string            `json:"language,omitempty"`
	Uri              string            `json:"uri,omitempty"`
	Url              string            `json:"url,omitempty"`
	Content          string            `json:"content"`
	Sensitive        bool              `json:"sensitive"`
	Account          Account           `json:"account"`
	Tags             []Tag             `json:"tags"`
	MediaAttachments []MediaAttachment `json:"media_attachments"`
}

type Account struct {
	Acct           string `json:"acct"`
	Discoverable   bool   `json:"discoverable"`
	DisplayName    string `json:"display_name"`
	Indexable      bool   `json:"indexable"`
	Note           string `json:"note"`
	Uri            string `json:"uri"`
	Url            string `json:"url"`
	FollowersCount int    `json:"followers_count"`
	StatusesCount  int    `json:"statuses_count"`
}

type Tag struct {
	Name string `json:"name"`
}

type MediaAttachment struct {
	Type       string `json:"type"`
	Url        string `json:"url"`
	PreviewUrl string `json:"preview_url"`
}
