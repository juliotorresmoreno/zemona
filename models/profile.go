package models

type Profile struct {
	ID        string `json:"id"         bson:"_id"`
	Username  string `json:"username"   bson:"username"`
	Name      string `json:"name"       bson:"name"`
	CreatedAt string `json:"created_at" bson:"created_at"`
}

func (p Profile) TableName() string {
	return "profiles"
}
