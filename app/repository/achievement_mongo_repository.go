package repository

import (
	"context"
	"errors"
	"time"

	"achievements-uas/app/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		collection: db.Collection("achievements"),
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

/*
=====================================================
FIND BY ID (OBJECT ID)
=====================================================
*/
func (r *AchievementMongoRepository) FindByID(
	ctx context.Context,
	id primitive.ObjectID,
) (*models.Achievement, error) {

	var result models.Achievement
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("achievement not found")
	}
	return &result, err
}

/*
=====================================================
WRAPPER FIND BY STRING ID
=====================================================
*/
func (r *AchievementMongoRepository) GetByID(
	ctx context.Context,
	id string,
) (*models.Achievement, error) {

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid achievement id")
	}
	return r.FindByID(ctx, objID)
}

/*
=====================================================
UPDATE
SUPPORT:
1) Update(ctx, *models.Achievement)
2) Update(ctx, primitive.ObjectID, bson.M)
=====================================================
*/
func (r *AchievementMongoRepository) Update(
	ctx context.Context,
	args ...interface{},
) error {

	// CASE 1
	if len(args) == 1 {
		a, ok := args[0].(*models.Achievement)
		if !ok {
			return errors.New("invalid argument for Update")
		}

		a.UpdatedAt = time.Now()
		update := bson.M{"$set": a}

		_, err := r.collection.UpdateByID(ctx, a.ID, update)
		return err
	}

	// CASE 2
	if len(args) == 2 {
		id, ok1 := args[0].(primitive.ObjectID)
		update, ok2 := args[1].(bson.M)

		if !ok1 || !ok2 {
			return errors.New("invalid arguments for Update")
		}

		if update["$set"] == nil {
			update["$set"] = bson.M{}
		}
		update["$set"].(bson.M)["updatedAt"] = time.Now()

		_, err := r.collection.UpdateByID(ctx, id, update)
		return err
	}

	return errors.New("invalid number of arguments for Update")
}

/*
=====================================================
QUERY BY STUDENT ID
=====================================================
*/
func (r *AchievementMongoRepository) FindByStudentID(
	ctx context.Context,
	studentID string,
) ([]models.Achievement, error) {

	cursor, err := r.collection.Find(ctx, bson.M{"studentId": studentID})
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
func (r *AchievementMongoRepository) SoftDelete(
	ctx context.Context,
	id primitive.ObjectID,
) error {

	update := bson.M{
		"$set": bson.M{
			"status":    "deleted",
			"updatedAt": time.Now(),
		},
	}
	_, err := r.collection.UpdateByID(ctx, id, update)
	return err
}
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
FR-010: FIND BY ACHIEVEMENT IDS (ADMIN)
=====================================================
*/
func (r *AchievementMongoRepository) FindByIDs(
	ctx context.Context,
	ids []primitive.ObjectID,
) ([]models.Achievement, error) {

	filter := bson.M{"_id": bson.M{"$in": ids}}

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
// =====================================================
// FIND BY IDS + FILTER + SORT
// =====================================================
func (r *AchievementMongoRepository) FindByIDsWithFilter(ctx context.Context,ids []primitive.ObjectID,sortBy string,order int,dateFrom, dateTo time.Time,
) ([]models.Achievement, error) {

	filter := bson.M{"_id": bson.M{"$in": ids}}

	if !dateFrom.IsZero() || !dateTo.IsZero() {
		filter["createdAt"] = bson.M{}
		if !dateFrom.IsZero() {
			filter["createdAt"].(bson.M)["$gte"] = dateFrom
		}
		if !dateTo.IsZero() {
			filter["createdAt"].(bson.M)["$lte"] = dateTo
		}
	}

	opts := options.Find()
	if sortBy != "" {
		opts.SetSort(bson.D{{Key: sortBy, Value: order}})
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
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

// =====================================================
// FIND ALL (REPORT / ADMIN)
// =====================================================
func (r *AchievementMongoRepository) FindAll(ctx context.Context,) ([]models.Achievement, error) {

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
