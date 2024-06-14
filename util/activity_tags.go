package util

// ActivityTags that the activitypub library fails to deserialize
type ActivityTags struct {
	Tag []ActivityTag `json:"tag,omitempty"`
}

type ActivityTag struct {
	Name string `json:"name"`
}
