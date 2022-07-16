package models

type QueryProfileInformation struct {
	ID    string `json:"_id" bson:"_id"`
	Count int64  `json:"count"`
}

func (t QueryProfileInformation) TableName() string {
	return "queryProfileInformation"
}
