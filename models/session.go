package models

type Session struct {
	Profile *Profile `json:"profile"`
	Token   string   `json:"token"`
}
