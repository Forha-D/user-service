package repository

import (
	"context"

	appErr "user-service/internal/errors"
	"user-service/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(col *mongo.Collection) *UserRepository {
	return &UserRepository{
		collection: col,
	}
}

// Find user by email
func (r *UserRepository) FindByEmail(email string) (*model.User, error) {

	var user model.User

	err := r.collection.FindOne(
		context.Background(),
		bson.M{"email": email},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, appErr.ErrUserNotFound
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(user *model.User) error {

	_, err := r.collection.InsertOne(
		context.Background(),
		user,
	)

	if mongo.IsDuplicateKeyError(err) {
		return appErr.ErrUserAlreadyExists
	}
	return err
}

func (r *UserRepository) UpdateUser(email string, update bson.M) error {

	res, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"email": email},
		bson.M{"$set": update},
	)

	if res.MatchedCount == 0 {
		return appErr.ErrUserNotFound
	}

	return err
}
