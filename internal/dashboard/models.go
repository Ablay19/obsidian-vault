package dashboard

import "time"

type User struct {
	ID           int64
	Username     string
	FirstName    string
	LastName     string
	LanguageCode string
	CreatedAt    time.Time
}
