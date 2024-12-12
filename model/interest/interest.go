package interest

import (
	"time"
)

type Data struct {

	// Description is a human-readable subscription description.
	Description string

	// Enabled defines whether subscription is active and may be used to deliver a message.
	Enabled bool

	// Expires defines a deadline when subscription becomes disabled regardless the Enabled value.
	Expires time.Time

	Created time.Time

	Updated time.Time

	Public bool

	Followers int64
}
