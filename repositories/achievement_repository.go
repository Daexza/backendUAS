package repository

import (
	"context"
	"database/sql"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"achievements-uas/models"
)

type AchievementListRepository struct {
	PG *sql.DB
	MG *mongo.Database
}

func NewAchievementListRepository(pg *sql.DB, mg *mongo.Database) *AchievementListRepository {
	return &AchievementListRepository{PG: pg, MG: mg}
}

func (r *AchievementListRepository) List(ctx context.Context, role string, userID string) ([]models.Achievement, error) {

	var studentIDs []string

	// Role: Mahasiswa → prestasi miliknya sendiri
	if role == "Mahasiswa" {
		row := r.PG.QueryRow(`SELECT id FROM students WHERE user_id=$1`, userID)
		var sid string
		if err := row.Scan(&sid); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, sid)
	}

	// Role: Dosen Wali → prestasi mahasiswa bimbingannya
	if role == "Dosen Wali" {
		row := r.PG.QueryRow(`SELECT id FROM lecturers WHERE user_id=$1`, userID)
		var lecturerID string
		if err := row.Scan(&lecturerID); err != nil {
			return nil, err
		}

		rows, err := r.PG.Query(`SELECT id FROM students WHERE advisor_id=$1`, lecturerID)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var id string
			rows.Scan(&id)
			studentIDs = append(studentIDs, id)
		}
	}

	// Role: Admin → semua prestasi
	if role == "Admin" {
		rows, err := r.PG.Query(`SELECT id FROM students`)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var id string
			rows.Scan(&id)
			studentIDs = append(studentIDs, id)
		}
	}

	// Ambil mongo_achievement_id dari achievement_references
	rows, err := r.PG.Query(`
		SELECT mongo_achievement_id 
		FROM achievement_references 
		WHERE student_id = ANY($1)
	`, pqArray(studentIDs))
	if err != nil {
		return nil, err
	}

	var mongoIDs []string
	for rows.Next() {
		var mid string
		rows.Scan(&mid)
		mongoIDs = append(mongoIDs, mid)
	}

	if len(mongoIDs) == 0 {
		return []models.Achievement{}, nil
	}

	// Ambil detail prestasi dari MongoDB
	col := r.MG.Collection("achievements")
	cur, err := col.Find(ctx, bson.M{
		"_id": bson.M{"$in": mongoIDs},
	})
	if err != nil {
		return nil, err
	}

	var results []models.Achievement
	for cur.Next(ctx) {
		var a models.Achievement
		cur.Decode(&a)
		results = append(results, a)
	}

	return results, nil
}

// helper
func pqArray(x []string) interface{} { return x }
