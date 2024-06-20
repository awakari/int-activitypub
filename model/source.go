package model

import "time"

type Source struct {
	ActorId  string
	GroupId  string
	UserId   string
	Type     string
	Name     string
	Summary  string
	Accepted bool
	Rejected bool
	Last     time.Time
	Created  time.Time
	SubId    string
	Term     string
	Err      string
}
