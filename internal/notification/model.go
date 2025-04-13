package notification

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	ID       primitive.ObjectID     `json:"id,omitempty" bson:"_id,omitempty"`
	Category string                 `json:"category" bson:"category"`
	Subtype  string                 `json:"subtype" bson:"subtype"`
	UserID   uint32                 `json:"userID" bson:"userID"`
	Payload  map[string]interface{} `json:"payload" bson:"payload"`
}
