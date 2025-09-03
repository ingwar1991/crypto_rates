package mongo_helper

import (
	"time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)


type User struct {
    ID primitive.ObjectID `bson:"_id,omitempty"`

    Email string `bson:"email"`
    APIKey string `bson:"api_key"`

    CreatedAt time.Time `bson:"created_at"`
}


type LoginCode struct {
    ID primitive.ObjectID `bson:"_id,omitempty"`
    
    Email string `bson:"email"`
    CodeHash string `bson:"code_hash"`

    CreatedAt time.Time `bson:"created_at"`
    ExpiresAt time.Time `bson:"expires_at"`
}


type ActiveSecret struct {
    ID primitive.ObjectID `bson:"_id,omitempty"`

    Email string `bson:"email"`
    Secret string `bson:"secret"`

    CreatedAt time.Time `bson:"created_at"`
    ExpiresAt time.Time `bson:"expires_at"`
}

func (s *ActiveSecret) IsValid(tNow time.Time) bool {
	if tNow.IsZero() {
		tNow = time.Now()
	}

	return time.Now().Compare(s.ExpiresAt) < 0 
}


type LogEntry struct {
    ID primitive.ObjectID `bson:"_id,omitempty"`

    Email string `bson:"email" json:"email"`
    Secret string `bson:"secret" json:"secret"`

    Endpoint string `bson:"endpoint" json:"endpoint"`
    Params map[string]any `bson:"params" json:"params,omitempty"`
    ResponseStatus int `bson:"response_status" json:"response_status"`

    Timestamp int64 `bson:"timestamp" json:"timestamp"`
}
