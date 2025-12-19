package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//
// ====================================================
// ACHIEVEMENT (MONGODB)
// ====================================================
//
type Achievement struct {
	ID          primitive.ObjectID   `bson:"_id,omitempty" json:"id,omitempty"`
	StudentID   string               `bson:"studentId" json:"studentid"`
	AchievementType string           `bson:"achievementType" json:"achievementType"`
	Title       string               `bson:"title" json:"title"`
	Description string               `bson:"description" json:"description"`
	Details     AchievementDetails   `bson:"details,omitempty" json:"details,omitempty"`
	Attachments []Attachment         `bson:"attachments,omitempty" json:"attachments,omitempty"`
	Tags        []string             `bson:"tags,omitempty" json:"tags,omitempty"`
	Points      int                  `bson:"points,omitempty" json:"points,omitempty"`
	Status      string               `bson:"status" json:"status"`
	CreatedAt   time.Time            `bson:"createdAt" json:"created_at"`
	UpdatedAt   time.Time            `bson:"updatedAt" json:"updated_at"`
	History     []AchievementHistory `bson:"history,omitempty" json:"history,omitempty"`
}

//
// ====================================================
// ACHIEVEMENT DETAILS (DYNAMIC FIELDS)
// ====================================================
//
type AchievementDetails struct {

	// ---------------------
	// Competition
	// ---------------------
	CompetitionName  string `bson:"competitionName,omitempty" json:"competition_name,omitempty"`
	CompetitionLevel string `bson:"competitionLevel,omitempty" json:"competition_level,omitempty"`
	Rank             int    `bson:"rank,omitempty" json:"rank,omitempty"`
	MedalType        string `bson:"medalType,omitempty" json:"medal_type,omitempty"`

	// ---------------------
	// Publication
	// ---------------------
	PublicationType  string   `bson:"publicationType,omitempty" json:"publication_type,omitempty"`
	PublicationTitle string   `bson:"publicationTitle,omitempty" json:"publication_title,omitempty"`
	Authors          []string `bson:"authors,omitempty" json:"authors,omitempty"`
	Publisher        string   `bson:"publisher,omitempty" json:"publisher,omitempty"`
	ISSN             string   `bson:"issn,omitempty" json:"issn,omitempty"`

	// ---------------------
	// Organization
	// ---------------------
	OrganizationName string  `bson:"organizationName,omitempty" json:"organization_name,omitempty"`
	Position         string  `bson:"position,omitempty" json:"position,omitempty"`
	Period           *Period `bson:"period,omitempty" json:"period,omitempty"`

	// ---------------------
	// Certification
	// ---------------------
	CertificationName   string    `bson:"certificationName,omitempty" json:"certification_name,omitempty"`
	IssuedBy            string    `bson:"issuedBy,omitempty" json:"issued_by,omitempty"`
	CertificationNumber string    `bson:"certificationNumber,omitempty" json:"certification_number,omitempty"`
	ValidUntil          time.Time `bson:"validUntil,omitempty" json:"valid_until,omitempty"`

	// ---------------------
	// General
	// ---------------------
	EventDate    time.Time              `bson:"eventDate,omitempty" json:"event_date,omitempty"`
	Location     string                 `bson:"location,omitempty" json:"location,omitempty"`
	Organizer    string                 `bson:"organizer,omitempty" json:"organizer,omitempty"`
	Score        float64                `bson:"score,omitempty" json:"score,omitempty"`
	CustomFields map[string]interface{} `bson:"customFields,omitempty" json:"custom_fields,omitempty"`
}

//
// ====================================================
// PERIOD (ORGANIZATION)
// ====================================================
//
type Period struct {
	Start time.Time `bson:"start,omitempty" json:"start,omitempty"`
	End   time.Time `bson:"end,omitempty" json:"end,omitempty"`
}

//
// ====================================================
// ATTACHMENT
// ====================================================
//
type Attachment struct {
	FileName   string    `bson:"fileName" json:"file_name"`
	FileURL    string    `bson:"fileUrl" json:"file_url"`
	FileType   string    `bson:"fileType" json:"file_type"`
	UploadedAt time.Time `bson:"uploadedAt" json:"uploaded_at"`
}

//
// ====================================================
// ACHIEVEMENT HISTORY
// ====================================================
//
type AchievementHistory struct {
	AchievementID string `bson:"achievementId,omitempty" json:"achievement_id,omitempty"`
	StudentID     string `bson:"studentId,omitempty" json:"student_id,omitempty"`
	Status        string    `bson:"status" json:"status"`
	ChangedBy     string    `bson:"changedBy" json:"changed_by"`
	ChangedAt     time.Time `bson:"changedAt" json:"changed_at"`
	Notes         string    `bson:"notes,omitempty" json:"notes,omitempty"`
}

//
// ====================================================
// ACHIEVEMENT REFERENCE (POSTGRESQL)
// ====================================================
//
type AchievementReference struct {
	ID                 string     `json:"id"`
	StudentID          string     `json:"student_id"`
	MongoAchievementID string     `json:"mongo_achievement_id"`
	Status             string     `json:"status"`
	SubmittedAt        *time.Time `json:"submitted_at,omitempty"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	VerifiedBy         *string    `json:"verified_by,omitempty"`
	RejectionNote      *string    `json:"rejection_note,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}
