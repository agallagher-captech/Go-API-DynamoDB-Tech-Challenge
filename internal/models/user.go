package models

type User struct {
	DynamoDBBase
	ID       UUID   `dynamodbav:"user_id"`
	Name     string `dynamodbav:"name"`
	Email    string `dynamodbav:"email"`
	Password string `dynamodbav:"password"`
}
