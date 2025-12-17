package repository

import (
	"context"
	"database/sql"
	"time"
	"strconv"
	"achievements-uas/app/models"
)

/*
=====================================================
STRUCT REPOSITORY (NO INTERFACE)
=====================================================
*/
type AchievementPostgresRepository struct {
	db *sql.DB
}

/*
=====================================================
CONSTRUCTOR
=====================================================
*/
func NewAchievementPostgresRepository(db *sql.DB) *AchievementPostgresRepository {
	return &AchievementPostgresRepository{db: db}
}

/*
=====================================================
FR-003: Create reference saat prestasi dibuat
=====================================================
*/
func (r *AchievementPostgresRepository) Create(
	ctx context.Context,
	ref *models.AchievementReference,
) error {

	query := `
		INSERT INTO achievement_references
		(id, student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		ref.ID,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Status,
		time.Now(),
		time.Now(),
	)
	return err
}

/*
=====================================================
FR-004: Update status (draft → submitted)
=====================================================
*/
func (r *AchievementPostgresRepository) UpdateStatus(
	ctx context.Context,
	mongoID,
	status string,
) error {

	query := `
		UPDATE achievement_references
		SET status=$1, updated_at=$2
		WHERE mongo_achievement_id=$3
	`

	_, err := r.db.ExecContext(ctx, query, status, time.Now(), mongoID)
	return err
}

/*
=====================================================
FR-007: Verify prestasi
=====================================================
*/
func (r *AchievementPostgresRepository) SetVerified(
	ctx context.Context,
	mongoID,
	dosenID string,
) error {

	query := `
		UPDATE achievement_references
		SET status='verified',
		    verified_by=$1,
		    verified_at=$2,
		    updated_at=$2
		WHERE mongo_achievement_id=$3
	`

	_, err := r.db.ExecContext(ctx, query, dosenID, time.Now(), mongoID)
	return err
}

/*
=====================================================
FR-008: Reject prestasi
=====================================================
*/
func (r *AchievementPostgresRepository) SetRejected(
	ctx context.Context,
	mongoID,
	note string,
) error {

	query := `
		UPDATE achievement_references
		SET status='rejected',
		    rejection_note=$1,
		    updated_at=$2
		WHERE mongo_achievement_id=$3
	`

	_, err := r.db.ExecContext(ctx, query, note, time.Now(), mongoID)
	return err
}

/*
=====================================================
FR-006: Ambil reference prestasi mahasiswa bimbingan
=====================================================
*/
func (r *AchievementPostgresRepository) FindByStudentIDs(
	ctx context.Context,
	studentIDs []string,
) ([]models.AchievementReference, error) {

	query := `
		SELECT id, student_id, mongo_achievement_id, status,
		       created_at, updated_at
		FROM achievement_references
		WHERE student_id = ANY($1)
	`

	rows, err := r.db.QueryContext(ctx, query, studentIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		if err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, ref)
	}
	return results, nil
}
/*
=====================================================
FR-010: ADMIN - GET ALL ACHIEVEMENTS (REFERENCE)
=====================================================
*/
func (r *AchievementPostgresRepository) GetAll(
	ctx context.Context,
	status string,
	limit, offset int,
) ([]models.AchievementReference, error) {

	query := `
		SELECT id, student_id, mongo_achievement_id, status,
		       created_at, updated_at
		FROM achievement_references
		WHERE ($1 = '' OR status = $1)
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.AchievementReference
	for rows.Next() {
		var ref models.AchievementReference
		if err := rows.Scan(
			&ref.ID,
			&ref.StudentID,
			&ref.MongoAchievementID,
			&ref.Status,
			&ref.CreatedAt,
			&ref.UpdatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, ref)
	}
	return results, nil
}
// =====================================================
// FR-010: GET ALL REFERENCES (ADMIN, PAGINATION)
// =====================================================
func (r *AchievementPostgresRepository) GetAllWithCount(
	ctx context.Context,
	status string,
	studentID string,
	limit, offset int,
) ([]models.AchievementReference, int, error) {

	where := "WHERE 1=1"
	args := []interface{}{}
	i := 1

	if status != "" {
		where += " AND status=$" + strconv.Itoa(i)
		args = append(args, status)
		i++
	}

	if studentID != "" {
		where += " AND student_id=$" + strconv.Itoa(i)
		args = append(args, studentID)
		i++
	}

	// count
	countQuery := `
		SELECT COUNT(*)
		FROM achievement_references
	` + where

	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// data
	dataQuery := `
		SELECT id, student_id, mongo_achievement_id, status,
		       created_at, updated_at
		FROM achievement_references
	` + where + `
		ORDER BY created_at DESC
		LIMIT $` + strconv.Itoa(i) + `
		OFFSET $` + strconv.Itoa(i+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []models.AchievementReference
	for rows.Next() {
		var r models.AchievementReference
		if err := rows.Scan(
			&r.ID,
			&r.StudentID,
			&r.MongoAchievementID,
			&r.Status,
			&r.CreatedAt,
			&r.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}

	return results, total, nil
}
