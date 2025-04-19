package notification

type AddNotification struct {
	Category string                 `json:"category" bson:"category"`
	Subtype  string                 `json:"subtype" bson:"subtype"`
	UserID   uint32                 `json:"userID" bson:"userID"`
	Payload  map[string]interface{} `json:"payload" bson:"payload"`
}
