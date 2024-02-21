package http

import (
	"encoding/json"
	vocab "github.com/go-ap/activitypub"
)

func FixContext(obj vocab.ActivityObject) (m map[string]any) {
	d, _ := json.Marshal(obj)
	m = make(map[string]any)
	_ = json.Unmarshal(d, &m)
	c, ok := m["context"]
	if ok {
		m["@context"] = c
		delete(m, "context")
	}
	return
}
