package repository

import (
	"achievements-uas/app/models"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementMongoRepository struct {
	Collection *mongo.Collection
}

func NewAchievementMongoRepository(c *mongo.Collection) *AchievementMongoRepository {
	return &AchievementMongoRepository{Collection: c}
}

// Convert string → ObjectID
func toObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}

// CREATE
func (r *AchievementMongoRepository) Create(a *models.Achievement) error {
	a.ID = primitive.NewObjectID()
	a.Status = "draft"
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()

	_, err := r.Collection.InsertOne(context.Background(), a)
	return err
}

// GET BY ID
func (r *AchievementMongoRepository) GetByID(id string) (*models.Achievement, error) {
	oid, err := toObjectID(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": oid, "deleted": bson.M{"$ne": true}}

	var a models.Achievement
	err = r.Collection.FindOne(context.Background(), filter).Decode(&a)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// UPDATE
func (r *AchievementMongoRepository) Update(a *models.Achievement) error {
	filter := bson.M{"_id": a.ID, "deleted": bson.M{"$ne": true}}
	a.UpdatedAt = time.Now()

	update := bson.M{"$set": a}
	_, err := r.Collection.UpdateOne(context.Background(), filter, update)
	return err
}

// SOFT DELETE
func (r *AchievementMongoRepository) SoftDelete(id string, studentID string) error {
	oid, err := toObjectID(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": oid, "studentId": studentID}
	update := bson.M{"$set": bson.M{"deleted": true, "updatedAt": time.Now()}}

	_, err = r.Collection.UpdateOne(context.Background(), filter, update)
	return err
}

// LIST (by filter)
func (r *AchievementMongoRepository) List(filter bson.M) ([]models.Achievement, error) {
	filter["deleted"] = bson.M{"$ne": true}

	cursor, err := r.Collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var list []models.Achievement
	for cursor.Next(context.Background()) {
		var a models.Achievement
		if err := cursor.Decode(&a); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

// SUBMIT
func (r *AchievementMongoRepository) Submit(id string, studentID string) error {
	oid, err := toObjectID(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": oid, "studentId": studentID, "status": "draft"}
	update := bson.M{"$set": bson.M{"status": "submitted", "updatedAt": time.Now()}}

	_, err = r.Collection.UpdateOne(context.Background(), filter, update)
	return err
}

// VERIFY
func (r *AchievementMongoRepository) Verify(id string) error {
	oid, err := toObjectID(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid, "status": "submitted"}
	update := bson.M{"$set": bson.M{"status": "verified", "updatedAt": time.Now()}}

	_, err = r.Collection.UpdateOne(context.Background(), filter, update)
	return err
}

// REJECT
func (r *AchievementMongoRepository) Reject(id string, note string) error {
	oid, err := toObjectID(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": oid, "status": "submitted"}
	update := bson.M{"$set": bson.M{
		"status":        "rejected",
		"rejectionNote": note,
		"updatedAt":     time.Now(),
	}}

	_, err = r.Collection.UpdateOne(context.Background(), filter, update)
	return err
}

// ATTACHMENT
func (r *AchievementMongoRepository) AddAttachment(id string, att models.Attachment) error {
	oid, err := toObjectID(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$push": bson.M{"attachments": att},
		"$set": bson.M{"updatedAt": time.Now()},
	}

	_, err = r.Collection.UpdateByID(context.Background(), oid, update)
	return err
}
