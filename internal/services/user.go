package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/agallagher-captech/blog/internal/models"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
)

type dynamoClient interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	// Add any other methods you might need from the DynamoDB client
}

var ErrNotFound = fmt.Errorf("item not found")
var ErrAlreadyExists = fmt.Errorf("item already exists")

// UsersService is a service capable of performing CRUD operations for
// models.User models.
type UsersService struct {
	logger *slog.Logger
	client dynamoClient
}

// NewUsersService creates a new UsersService and returns a pointer to it.
func NewUsersService(logger *slog.Logger, client dynamoClient) *UsersService {
	return &UsersService{
		logger: logger,
		client: client,
	}
}

// CreateUser attempts to create the provided user, returning a fully hydrated
// models.User or an error.
// CreateUser attempts to create a new user in the database. It returns an error if the user could not be created.
func (s *UsersService) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	s.logger.InfoContext(ctx, "Creating user", "id", user.ID)

	// Check if the user already exists
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("BlogContent"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: fmt.Sprintf("USER#%s", user.ID.String()),
			},
			"SK": &types.AttributeValueMemberS{
				Value: "PROFILE",
			},
		},
	})
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.CreateUser] failed to check if user exists: %w",
			err,
		)
	}

	// If the user already exists, return an error
	if result.Item != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.CreateUser] user already exists.",
		)
	}

	// Marshal the user struct into a map of DynamoDB AttributeValues
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.CreateUser] failed to marshal user: %w",
			err,
		)
	}

	// Add the PK and SK to the item
	item["PK"] = &types.AttributeValueMemberS{
		Value: fmt.Sprintf("USER#%s", user.ID.String()),
	}
	item["SK"] = &types.AttributeValueMemberS{
		Value: "PROFILE",
	}

	// Put the item into DynamoDB
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String("BlogContent"),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	})
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.CreateUser] failed to put item: %w",
			err,
		)
	}

	return user, nil
}

// ReadUser attempts to read a user from the database using the provided id. A
// fully hydrated models.User or error is returned.
func (s *UsersService) ReadUser(ctx context.Context, id uuid.UUID) (models.User, error) {
	s.logger.InfoContext(ctx, "Reading user", "id", id)

	// get item from DynamoDB by PK and SK
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("BlogContent"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: fmt.Sprintf("USER#%s", id.String()),
			},
			"SK": &types.AttributeValueMemberS{
				Value: "PROFILE",
			},
		},
	})
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.ReadUser] failed to get item: %w",
			err,
		)
	}

	// handle item not found
	if result.Item == nil {
		return models.User{}, ErrNotFound
	}

	// Unmarshal the results into the models.User struct
	var user models.User
	if err = attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.ReadUser] failed to unmarshal result: %w",
			err,
		)
	}

	return user, nil
}

// UpdateUser attempts to perform an update of the user with the provided id,
// updating it to reflect the properties on the provided patch object. A
// models.User or an error is returned.
func (s *UsersService) UpdateUser(ctx context.Context, id uuid.UUID, patch models.User) (models.User, error) {
	s.logger.InfoContext(ctx, "Updating user", "id", id)

	// Check if the user exists
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String("BlogContent"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: fmt.Sprintf("USER#%s", id.String()),
			},
			"SK": &types.AttributeValueMemberS{
				Value: "PROFILE",
			},
		},
	})
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.UpdateUser] failed to get item: %w",
			err,
		)
	}

	// Handle item not found
	if result.Item == nil {
		return models.User{}, ErrNotFound
	}

	// Unmarshal the existing user into the models.User struct
	var existingUser models.User
	if err = attributevalue.UnmarshalMap(result.Item, &existingUser); err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.UpdateUser] failed to unmarshal result: %w",
			err,
		)
	}

	// Update the existing user with the patch data
	if patch.Name != "" {
		existingUser.Name = patch.Name
	}
	if patch.Email != "" {
		existingUser.Email = patch.Email
	}
	if patch.Password != "" {
		existingUser.Password = patch.Password
	}

	// Marshal the updated user struct into a map of DynamoDB AttributeValues
	updatedItem, err := attributevalue.MarshalMap(existingUser)
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.UpdateUser] failed to marshal updated user: %w",
			err,
		)
	}

	// Update the item in DynamoDB
	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String("BlogContent"),
		Item:      updatedItem,
	})
	if err != nil {
		return models.User{}, fmt.Errorf(
			"[in main.UsersService.UpdateUser] failed to put updated item: %w",
			err,
		)
	}

	return existingUser, nil
}

// DeleteUser attempts to delete the user with the provided id. An error is
// returned if the delete fails.
func (s *UsersService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	s.logger.InfoContext(ctx, "Deleting user", "id", id)

	// Perform the delete operation in DynamoDB
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String("BlogContent"),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{
				Value: fmt.Sprintf("USER#%s", id.String()),
			},
			"SK": &types.AttributeValueMemberS{
				Value: "PROFILE",
			},
		},
		ConditionExpression: aws.String("attribute_exists(PK) AND attribute_exists(SK)"),
	})
	if err != nil {
		return fmt.Errorf(
			"[in main.UsersService.DeleteUser] failed to delete item: %w",
			err,
		)
	}

	return nil
}

// ListUsers attempts to list all users in the database using a GSI. A slice of models.User
// or an error is returned.
func (s *UsersService) ListUsers(ctx context.Context) ([]models.User, error) {
	s.logger.InfoContext(ctx, "Listing all users")

	// Define the input for the Query operation
	input := &dynamodb.QueryInput{
		TableName:              aws.String("BlogContent"),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{
				Value: "USER",
			},
		},
	}

	// Perform the Query operation
	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf(
			"[in main.UsersService.ListUsers] failed to query items: %w",
			err,
		)
	}

	// Unmarshal the results into a slice of models.User
	var users []models.User
	if err = attributevalue.UnmarshalListOfMaps(result.Items, &users); err != nil {
		return nil, fmt.Errorf(
			"[in main.UsersService.ListUsers] failed to unmarshal result: %w",
			err,
		)
	}

	return users, nil
}
