package util

// ActivityTags that the activitypub library fails to deserialize
type ActivityTags struct {
	Tag    []ActivityTag `json:"tag,omitempty"`
	Object ObjectTags    `json:"object,omitempty"`
}

type ObjectTags struct {
	Tag []ActivityTag `json:"tag,omitempty"`
}

type ActivityTag struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ActivityContentMap struct {
	ContentMap map[string]string `json:"contentMap,omitempty"`
}
