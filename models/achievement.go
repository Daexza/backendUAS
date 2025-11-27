package models

import "time"

type Achievement struct {
	ID              string                 `json:"id" bson:"_id,omitempty"`
	StudentID       string                 `json:"studentId" bson:"studentId"`
	AchievementType string                 `json:"achievementType" bson:"achievementType"`
	Title           string                 `json:"title" bson:"title"`
	Description     string                 `json:"description" bson:"description"`
	Details         map[string]interface{} `json:"details" bson:"details"`
	Attachments     []Attachment           `json:"attachments" bson:"attachments"`
	Tags            []string               `json:"tags" bson:"tags"`
	Points          int                    `json:"points" bson:"points"`
	CreatedAt       time.Time              `json:"createdAt" bson:"createdAt"`
	UpdatedAt       time.Time              `json:"updatedAt" bson:"updatedAt"`
}

type Attachment struct {
	FileName   string    `json:"fileName" bson:"fileName"`
	FileURL    string    `json:"fileUrl" bson:"fileUrl"`
	FileType   string    `json:"fileType" bson:"fileType"`
	UploadedAt time.Time `json:"uploadedAt" bson:"uploadedAt"`
}
