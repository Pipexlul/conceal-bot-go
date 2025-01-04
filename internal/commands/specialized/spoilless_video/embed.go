package spoillessvideo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Metadata struct {
	Description string `bson:"description"`
	OGTitle     string `bson:"og_title"`
	OGURL       string `bson:"og_url"`
}

type Embed struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	VideoID     string             `bson:"video_id"`
	CustomTitle string             `bson:"custom_title"`
	Thumbnail   string             `bson:"-"`
	Metadata    Metadata           `bson:"metadata"`
	CreatedAt   time.Time          `bson:"created_at"`
}
