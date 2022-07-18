package models

type Profile struct {
	ID          string `json:"id"          bson:"_id"`
	Username    string `json:"username"    bson:"username"`
	Name        string `json:"name"        bson:"name"`
	ImageSrc    string `json:"image_src"   bson:"image_src"`
	Description string `json:"description" bson:"description"`
	CreatedAt   string `json:"created_at"  bson:"created_at"`
}

func (p Profile) TableName() string {
	return "profiles"
}
