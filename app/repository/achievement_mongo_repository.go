package repository

import (
	"context"
	"errors"
	"fmt"  // Diperlukan untuk fmt.Errorf
	"time"

	"achievements-uas/app/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

/*
=====================================================
STRUCT REPOSITORY (NO INTERFACE)
=====================================================
*/
type AchievementMongoRepository struct {
	collection *mongo.Collection
}

/*
=====================================================
CONSTRUCTOR
=====================================================
*/
func NewAchievementMongoRepository(db *mongo.Database) *AchievementMongoRepository {
	return &AchievementMongoRepository{
		collection: db.Collection("achievement"),
	}
}

/*
=====================================================
CREATE
=====================================================
*/
func (r *AchievementMongoRepository) Create(
	ctx context.Context,
	a *models.Achievement,
) (*models.Achievement, error) {

	a.ID = primitive.NewObjectID()
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *AchievementMongoRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Achievement, error) {
    var result models.Achievement
    // Tambahkan filter status $ne (not equal) deleted
    filter := bson.M{
        "_id":    id,
        "status": bson.M{"$ne": "deleted"},
    }
    
    err := r.collection.FindOne(ctx, filter).Decode(&result)
    if err == mongo.ErrNoDocuments {
        return nil, errors.New("achievement not found or has been deleted")
    }
    return &result, err
}
/*
=====================================================
WRAPPER GET BY STRING ID
=====================================================
*/
func (r *AchievementMongoRepository) GetByID(ctx context.Context, id string) (*models.Achievement, error) {
    // 1. Konversi string ID dari Postman ke ObjectID MongoDB
    objID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, errors.New("invalid id format")
    }

    var result models.Achievement
    // 2. Cari di database
    err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&result)
    if err != nil {
        return nil, err // Akan mengembalikan mongo.ErrNoDocuments
    }
    return &result, nil
}

/*
=====================================================
UPDATE (Fixed)
=====================================================
*/
func (r *AchievementMongoRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) error {
	// Menggunakan r.collection (bukan r.DB) sesuai dengan struct definition
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no document found with ID %s", id.Hex())
	}

	return nil
}
/*
=====================================================
QUERY BY STUDENT ID (Filtered by Soft Delete)
=====================================================
*/
func (r *AchievementMongoRepository) FindByStudentID(
	ctx context.Context,
	studentID string,
) ([]models.Achievement, error) {

	// Filter: Milik mahasiswa tertentu DAN belum dihapus
	filter := bson.M{
		"studentId": studentID,
		"status":    bson.M{"$ne": "deleted"},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Achievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if results == nil {
		results = []models.Achievement{}
	}

	return results, nil
}

/*
=====================================================
FIND BY MULTIPLE STUDENT IDS (Dosen Wali)
=====================================================
*/
func (r *AchievementMongoRepository) FindByStudentIDs(
	ctx context.Context,
	studentIDs []string,
) ([]models.Achievement, error) {

	filter := bson.M{"studentId": bson.M{"$in": studentIDs}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Achievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

/*
=====================================================
SOFT DELETE
=====================================================
*/
func (r *AchievementMongoRepository) SoftDelete(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$set": bson.M{
			"status":    "deleted",
			"updatedAt": time.Now(),
		},
	}
	// Menggunakan UpdateOne untuk mengubah status field tanpa menghapus dokumen
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

/*
=====================================================
ADD ATTACHMENT
=====================================================
*/
func (r *AchievementMongoRepository) AddAttachment(ctx context.Context,id primitive.ObjectID,attachment models.Attachment,
) error {

	update := bson.M{
		"$push": bson.M{
			"attachments": attachment,
		},
		"$set": bson.M{
			"updatedAt": time.Now(),
		},
	}

	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}


/*
=====================================================
FIND BY OBJECT IDS (Filtered by Soft Delete)
=====================================================
*/
func (r *AchievementMongoRepository) FindByIDs(ctx context.Context,ids []primitive.ObjectID,
) ([]models.Achievement, error) {

	// Filter: Mencari ID yang ada di dalam slice 'ids' 
	// DAN statusnya TIDAK SAMA DENGAN 'deleted'
	filter := bson.M{
		"_id":    bson.M{"$in": ids},
		"status": bson.M{"$ne": "deleted"},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Achievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Jika results kosong, return array kosong bukan nil agar JSON di Postman []
	if results == nil {
		results = []models.Achievement{}
	}

	return results, nil
}

/*
=====================================================
FIND ALL
=====================================================
*/
func (r *AchievementMongoRepository) FindAll(ctx context.Context) ([]models.Achievement, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []models.Achievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
func (r *AchievementPostgresRepository) GetByMongoID(ctx context.Context, mongoID string) (*models.AchievementReference, error) {
    query := `
        SELECT id, student_id, mongo_achievement_id, status, created_at, updated_at
        FROM achievement_references
        WHERE mongo_achievement_id = $1
    `
    var ref models.AchievementReference
    err := r.db.QueryRowContext(ctx, query, mongoID).Scan(
        &ref.ID,
        &ref.StudentID, // Ini adalah UUID Mahasiswa
        &ref.MongoAchievementID,
        &ref.Status,
        &ref.CreatedAt,
        &ref.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &ref, nil
}