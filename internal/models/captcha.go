package models

import "time"

type Captcha struct {
	ID        string    `json:"id"`
	Answer    string    `json:"answer"`
	ExpiresAt time.Time `json:"-"`
}
