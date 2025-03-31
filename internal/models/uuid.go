package models

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// UUID is a custom type that wraps an uuid.UUID and implements the Marshaler
// and Unmarshaler interface from the
// `github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue` package.
type UUID struct {
	uuid.UUID
}

// UnmarshalDynamoDBAttributeValue unmarshals a UUID from a DynamoDB
// types.AttributeValue. It implements the attributevalue.Marshaler interface.
func (u *UUID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	s, ok := av.(*types.AttributeValueMemberS)
	if !ok {
		return fmt.Errorf("expected AttributeValueMemberS, got %T", av)
	}

	id, err := uuid.Parse(s.Value)
	if err != nil {
		return err
	}

	*u = UUID{UUID: id}
	return nil
}

// MarshalDynamoDBAttributeValue marshals a UUID into a DynamoDB
// types.AttributeValue. It implements the attributevalue.Marshaler interface.
func (u *UUID) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: u.UUID.String()}, nil
}
