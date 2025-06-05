package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/agallagher-captech/blog/internal/models"
	"github.com/agallagher-captech/blog/internal/services/mock"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var testErr = errors.New("test error")

func TestUsersService_ReadUser(t *testing.T) {
	testcases := map[string]struct {
		mockCalled     bool
		mockInput      []any
		mockOutput     []any
		input          uuid.UUID
		expectedOutput models.User
		expectedError  error
	}{
		"happy path": {
			mockCalled: true,
			mockInput: []any{
				context.TODO(),
				&dynamodb.GetItemInput{
					TableName: aws.String("BlogContent"),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{
							Value: "USER#d2eddb69-f92f-694d-450d-e7cdb6decce3",
						},
						"SK": &types.AttributeValueMemberS{
							Value: "PROFILE",
						},
					},
				},
			},
			mockOutput: []any{
				&dynamodb.GetItemOutput{
					Item: map[string]types.AttributeValue{
						"email":    &types.AttributeValueMemberS{Value: "testUser@example.com"},
						"GSI1PK":   &types.AttributeValueMemberS{Value: "USER"},
						"user_id":  &types.AttributeValueMemberS{Value: "d2eddb69-f92f-694d-450d-e7cdb6decce3"},
						"GSI1SK":   &types.AttributeValueMemberS{Value: "USER#d2eddb69-f92f-694d-450d-e7cdb6decce3"},
						"SK":       &types.AttributeValueMemberS{Value: "PROFILE"},
						"PK":       &types.AttributeValueMemberS{Value: "USER#d2eddb69-f92f-694d-450d-e7cdb6decce3"},
						"name":     &types.AttributeValueMemberS{Value: "Test User"},
						"password": &types.AttributeValueMemberS{Value: "Test Password"},
					},
				},
				nil,
			},
			input: uuid.MustParse("d2eddb69-f92f-694d-450d-e7cdb6decce3"),
			expectedOutput: models.User{
				DynamoDBBase: models.DynamoDBBase{
					PK:     "USER#d2eddb69-f92f-694d-450d-e7cdb6decce3",
					SK:     "PROFILE",
					GSI1PK: "USER",
					GSI1SK: "USER#d2eddb69-f92f-694d-450d-e7cdb6decce3",
				},
				ID:       models.UUID{UUID: uuid.MustParse("d2eddb69-f92f-694d-450d-e7cdb6decce3")},
				Name:     "Test User",
				Email:    "testUser@example.com",
				Password: "Test Password",
			},
			expectedError: nil,
		},
		"user not found": {
			mockCalled: true,
			mockInput: []any{
				context.TODO(),
				&dynamodb.GetItemInput{
					TableName: aws.String("BlogContent"),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{
							Value: "USER#00000000-0000-0000-0000-000000000000",
						},
						"SK": &types.AttributeValueMemberS{
							Value: "PROFILE",
						},
					},
				},
			},
			mockOutput: []any{
				&dynamodb.GetItemOutput{
					Item: nil, // Simulate no item found
				},
				errors.New("failed to get item: item not found"),
			},
			input:          uuid.MustParse("00000000-0000-0000-0000-000000000000"), // Non-existent user ID
			expectedOutput: models.User{},                                          // Expect an empty user object
			expectedError:  fmt.Errorf("failed to get item: %w", testErr),          // Replace with the actual error returned by your service
		},
	}
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			mockClient := new(mock.DynamoClient)
			logger := slog.Default()

			if tc.mockCalled {
				mockClient.
					On("GetItem", tc.mockInput...).
					Return(tc.mockOutput...).
					Once()
			}

			userService := UsersService{
				logger: logger,
				client: mockClient,
			}

			output, err := userService.ReadUser(context.TODO(), tc.input)

			// validate errors
			if tc.expectedError != nil {
				//assert.Equal(t, tc.expectedError.Error(), err, "errors did not match")
				assert.ErrorIs(t, err, testErr, "error did not wrap the expected error")
				assert.ErrorContains(t, err, testErr.Error(), "error did not contain the expected message")
				assert.ErrorContains(t, err, "failed to get item", "error did not contain the expected message")
			}
			assert.Equal(t, tc.expectedOutput, output, "returned data does not match")

			if tc.mockCalled {
				mockClient.AssertExpectations(t)
			} else {
				mockClient.AssertNotCalled(t, "GetItem")
			}
		})
	}
}
