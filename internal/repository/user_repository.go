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
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {

	var user model.User

	err := r.collection.FindOne(
		ctx,
		bson.M{"email": email},
	).Decode(&user)

	if err == mongo.ErrNoDocuments {
		return nil, appErr.ErrUserNotFound
	}

	return &user, err
}

func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {

	_, err := r.collection.InsertOne(
		ctx,
		user,
	)

	if mongo.IsDuplicateKeyError(err) {
		return appErr.ErrUserAlreadyExists
	}
	return err
}

func (r *UserRepository) UpdateUser(ctx context.Context, email string, update bson.M) error {

	res, err := r.collection.UpdateOne(
		ctx,
		bson.M{"email": email},
		bson.M{"$set": update},
	)

	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return appErr.ErrUserNotFound
	}

	return nil
}
