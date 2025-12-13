package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Achievement utama (MongoDB)
type Achievement struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	StudentID   string             `bson:"studentId"`         // UUID dari PostgreSQL
	Type        string             `bson:"achievementType"`   // academic, competition, organization, publication, certification, other
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	Details     AchievementDetails `bson:"details,omitempty"`
	Attachments []Attachment       `bson:"attachments,omitempty"`
	Tags        []string           `bson:"tags,omitempty"`
	Points      int                `bson:"points,omitempty"`
	Status      string             `bson:"status"` // draft, submitted
	CreatedAt   time.Time          `bson:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt"`
	History     []AchievementHistory `bson:"history,omitempty"` // histori perubahan status
}

// Detail prestasi
type AchievementDetails struct {
	// Competition
	CompetitionName  string `bson:"competitionName,omitempty"`
	CompetitionLevel string `bson:"competitionLevel,omitempty"`
	Rank             int    `bson:"rank,omitempty"`
	MedalType        string `bson:"medalType,omitempty"`

	// Publication
	PublicationType  string   `bson:"publicationType,omitempty"`
	PublicationTitle string   `bson:"publicationTitle,omitempty"`
	Authors          []string `bson:"authors,omitempty"`
	Publisher        string   `bson:"publisher,omitempty"`
	ISSN             string   `bson:"issn,omitempty"`

	// Organization
	OrganizationName string   `bson:"organizationName,omitempty"`
	Position         string   `bson:"position,omitempty"`
	Period           *Period  `bson:"period,omitempty"`

	// Certification
	CertificationName   string    `bson:"certificationName,omitempty"`
	IssuedBy            string    `bson:"issuedBy,omitempty"`
	CertificationNumber string    `bson:"certificationNumber,omitempty"`
	ValidUntil          time.Time `bson:"validUntil,omitempty"`

	// Field umum
	EventDate    time.Time              `bson:"eventDate,omitempty"`
	Location     string                 `bson:"location,omitempty"`
	Organizer    string                 `bson:"organizer,omitempty"`
	Score        float64                `bson:"score,omitempty"`
	CustomFields map[string]interface{} `bson:"customFields,omitempty"`
}

// Period untuk organisasi
type Period struct {
	Start time.Time `bson:"start,omitempty"`
	End   time.Time `bson:"end,omitempty"`
}

// Attachment
type Attachment struct {
	FileName   string    `bson:"fileName"`
	FileURL    string    `bson:"fileUrl"`
	FileType   string    `bson:"fileType"`
	UploadedAt time.Time `bson:"uploadedAt"`
}

// History prestasi
type AchievementHistory struct {
	AchievementID string    `bson:"achievementId"`   // Mongo Achievement ID
	StudentID     string    `bson:"studentId"`       // UUID mahasiswa
	Status        string    `bson:"status"`          // draft, submitted, verified, rejected
	ChangedBy     string    `bson:"changedBy"`       // UserID mahasiswa atau dosen
	ChangedAt     time.Time `bson:"changedAt"`       // waktu perubahan status
	Notes         string    `bson:"notes,omitempty"` // optional, misal catatan reject
}

// PostgreSQL Reference
type AchievementReference struct {
	ID                 string    `json:"id"`
	StudentID          string    `json:"student_id"`
	MongoAchievementID string    `json:"mongo_achievement_id"`
	Status             string    `json:"status"` // draft, submitted
	SubmittedAt        time.Time `json:"submitted_at"`
	VerifiedAt         time.Time `json:"verified_at"`
	VerifiedBy         string    `json:"verified_by"`
	RejectionNote      string    `json:"rejection_note"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}