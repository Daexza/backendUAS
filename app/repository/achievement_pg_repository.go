package repository

import (
	"database/sql"
	"time"
	"achievements-uas/app/models"
)

type AchievementPGRepository struct {
	DB *sql.DB
}

func NewAchievementPGRepository(db *sql.DB) *AchievementPGRepository {
	return &AchievementPGRepository{DB: db}
}

// CREATE REFERENCE
func (r *AchievementPGRepository) Create(ref *models.AchievementReference) error {
	ref.Status = "draft"
	ref.CreatedAt = time.Now()
	ref.UpdatedAt = time.Now()

	_, err := r.DB.Exec(`
		INSERT INTO achievement_references 
		(id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, ref.ID, ref.StudentID, ref.MongoAchievementID, ref.Status, ref.CreatedAt, ref.UpdatedAt)
	return err
}

// UPDATE STATUS (by mongo_id)
func (r *AchievementPGRepository) UpdateStatusByMongoID(mongoID, status, verifiedBy, note string) error {
	_, err := r.DB.Exec(`
		UPDATE achievement_references
		SET status=$2, verified_by=$3, rejection_note=$4, updated_at=$5
		WHERE mongo_achievement_id=$1
	`, mongoID, status, verifiedBy, note, time.Now())
	return err
}

// HISTORY
func (r *AchievementPGRepository) AddHistory(h *models.AchievementHistory) error {
	h.ChangedAt = time.Now()
	_, err := r.DB.Exec(`
		INSERT INTO achievement_histories (achievement_id, student_id, status, changed_by, changed_at, notes)
		VALUES ($1,$2,$3,$4,$5,$6)
	`, h.AchievementID, h.StudentID, h.Status, h.ChangedBy, h.ChangedAt, h.Notes)
	return err
}

func (r *AchievementPGRepository) GetHistoryByMongoID(id string) ([]models.AchievementHistory, error) {
	rows, err := r.DB.Query(`
		SELECT achievement_id, student_id, status, changed_by, changed_at, notes
		FROM achievement_histories
		WHERE achievement_id=$1
		ORDER BY changed_at ASC
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.AchievementHistory
	for rows.Next() {
		var h models.AchievementHistory
		rows.Scan(&h.AchievementID, &h.StudentID, &h.Status, &h.ChangedBy, &h.ChangedAt, &h.Notes)
		list = append(list, h)
	}
	return list, nil
}
