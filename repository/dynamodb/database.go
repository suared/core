package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awsDynamoDB "github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	_ "github.com/suared/core/v2/infra"
	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/security"
)

// BaseDAO provides standard DB fields that can be embedded in a custom user DAO.
// Note: For DynamoDB, Hash and Sort keys must be top-level attributes in the struct that is marshaled.
// They cannot be in the embedded Model.
type BaseDAO[M any] struct {
	Model     M         `dynamodbav:"model"`
	CreatedBy string    `dynamodbav:"createdBy,omitempty"`
	UpdatedBy string    `dynamodbav:"updatedBy,omitempty"`
	CreatedDt time.Time `dynamodbav:"createdDt,omitempty"`
	UpdatedDt time.Time `dynamodbav:"updatedDt,omitempty"`
}

// ModelObject defines an interface that model objects must implement.
type ModelObject interface {
	IsEmpty() bool
}

// DynamoSession - Implementation of DynamoSession
type DynamoSession struct {
	session *awsDynamoDB.DynamoDB
}

// Session - Return this session/ implement the Session interface
func (s *DynamoSession) Session() repository.Session {
	return s
}

func newDynamoSession(awsDynamoDBSession *awsDynamoDB.DynamoDB) *DynamoSession {
	return &DynamoSession{session: awsDynamoDBSession}
}

// ValidAction performs a security check using reflection to find the user field.
func ValidAction(ctx context.Context, action string, dao any) error {
	auth := security.GetAuth(ctx)
	authUser := auth.GetUser()

	val := reflect.ValueOf(dao)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var daoUser string
	if val.Kind() == reflect.Struct {
		// Look for CreatedBy, common in BaseDAO
		field := val.FieldByName("CreatedBy")
		if field.IsValid() && field.Kind() == reflect.String {
			daoUser = field.String()
		} else {
			// As a fallback, look for a UserID field
			field := val.FieldByName("UserID")
			if field.IsValid() && field.Kind() == reflect.String {
				daoUser = field.String()
			}
		}
	}

	if daoUser == "" {
		return nil
	}

	if authUser == daoUser || auth.IsAdmin() {
		return nil
	}

	return fmt.Errorf("security: User: %v does not have access to %v for %v", authUser, action, daoUser)
}

// CreateTable will create a dynamodb table if it doesn't exist.
func CreateTable(repo repository.Repository) (repository.Repository, error) {
	config := repo.Config().Values()
	table := config["table"]
	if table == "" {
		return nil, errors.New("DynamoDB table name cannot be empty")
	}

	backend := config["backend"]
	if backend != "dynamoDB" {
		return nil, fmt.Errorf("DynamoDB backend called for repository without the appropriate backend type: %v vs %v", backend, "dynamoDB")
	}

	region := config["region"]
	endpoint := config["endpoint"]
	hashKeyName := config["hashKeyName"]
	if hashKeyName == "" {
		return nil, errors.New("DynamoDB hashKeyName config is required")
	}
	sortKeyName := config["sortKeyName"]
	if sortKeyName == "" {
		return nil, errors.New("DynamoDB sortKeyName config is required")
	}

	rcu, err := strconv.ParseInt(config["rcu"], 10, 64)
	if err != nil {
		return nil, errors.New("rcu is not a number")
	}

	wcu, err := strconv.ParseInt(config["wcu"], 10, 64)
	if err != nil {
		return nil, errors.New("wcu is not a number")
	}

	env := config["env"]
	if env == "" || env == "dev" || env == "development" {
		env = "dev"
	}

	awsConfig := aws.Config{Region: aws.String(region), Endpoint: aws.String(endpoint)}
	awsDynamoSession := awsSession.Must(awsSession.NewSessionWithOptions(awsSession.Options{Config: awsConfig}))
	awsDynamoService := awsDynamoDB.New(awsDynamoSession)

	repo.SetSession(newDynamoSession(awsDynamoService))

	if env != "dev" {
		return repo, nil
	}

	req := &awsDynamoDB.DescribeTableInput{TableName: aws.String(table)}
	_, err = awsDynamoService.DescribeTable(req)

	if err != nil {
		awsTableCreateParams := &awsDynamoDB.CreateTableInput{
			AttributeDefinitions: []*awsDynamoDB.AttributeDefinition{
				{AttributeName: aws.String(hashKeyName), AttributeType: aws.String("S")},
				{AttributeName: aws.String(sortKeyName), AttributeType: aws.String("S")},
			},
			KeySchema: []*awsDynamoDB.KeySchemaElement{
				{AttributeName: aws.String(hashKeyName), KeyType: aws.String("HASH")},
				{AttributeName: aws.String(sortKeyName), KeyType: aws.String("RANGE")},
			},
			ProvisionedThroughput: &awsDynamoDB.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(rcu),
				WriteCapacityUnits: aws.Int64(wcu),
			},
			TableName: aws.String(table),
		}

		_, err = awsDynamoService.CreateTable(awsTableCreateParams)
		if err != nil {
			return nil, fmt.Errorf("failed to create table: %w", err)
		}
	}

	return repo, nil
}

// InsertOrUpdate is a generic method to insert or update a dynamo table.
func InsertOrUpdate(ctx context.Context, repo repository.Repository, item any) error {
	awsDAO, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal item: %w", err)
	}

	input := &awsDynamoDB.PutItemInput{
		Item:      awsDAO,
		TableName: aws.String(repo.Config().Values()["table"]),
	}

	dbSession := repo.Session().(*DynamoSession).session
	_, err = dbSession.PutItem(input)
	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}
	return nil
}

// Select returns a list of objects matching the hash key and optional filters.
func Select[D any](ctx context.Context, repo repository.Repository, hashKey string, strFilterVals ...string) ([]D, error) {
	config := repo.Config().Values()
	attrs := map[string]*awsDynamoDB.AttributeValue{
		":column": {S: aws.String(hashKey)},
	}

	var filterExpr *string
	if len(strFilterVals) > 0 {
		var filterBuilder strings.Builder
		countAttrs := len(strFilterVals) / 2
		for i := 0; i < countAttrs*2; i += 2 {
			if i > 0 {
				filterBuilder.WriteString(" AND ")
			}
			filterBuilder.WriteString(strFilterVals[i])

			valueExpr := strings.Split(strFilterVals[i+1], ":")
			attrName := ":" + valueExpr[0]
			attrType := valueExpr[1]
			attrValue := valueExpr[2]

			if attrType == "S" {
				attrs[attrName] = &awsDynamoDB.AttributeValue{S: aws.String(attrValue)}
			} else if attrType == "N" {
				attrs[attrName] = &awsDynamoDB.AttributeValue{N: aws.String(attrValue)}
			}
		}
		builtFilter := filterBuilder.String()
		filterExpr = &builtFilter
	}

	input := &awsDynamoDB.QueryInput{
		ExpressionAttributeValues: attrs,
		KeyConditionExpression:    aws.String(config["hashKeyName"] + " = :column"),
		FilterExpression:          filterExpr,
		TableName:                 aws.String(config["table"]),
	}

	dbSession := repo.Session().(*DynamoSession).session
	result, err := dbSession.Query(input)
	if err != nil {
		return nil, err
	}

	var daoResultList []D
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &daoResultList)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	return daoResultList, nil
}

// SelectOne returns a single object matching the hash and sort keys.
func SelectOne[D any](ctx context.Context, repo repository.Repository, hashKey, sortKey string) (D, error) {
	config := repo.Config().Values()
	input := &awsDynamoDB.QueryInput{
		ExpressionAttributeValues: map[string]*awsDynamoDB.AttributeValue{
			":thehash": {S: aws.String(hashKey)},
			":thesort": {S: aws.String(sortKey)},
		},
		KeyConditionExpression: aws.String(fmt.Sprintf("%s = :thehash AND %s = :thesort", config["hashKeyName"], config["sortKeyName"])),
		TableName:              aws.String(config["table"]),
	}

	dbSession := repo.Session().(*DynamoSession).session
	result, err := dbSession.Query(input)
	if err != nil {
		return *new(D), err
	}

	var daoResultList []D
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &daoResultList)
	if err != nil {
		return *new(D), fmt.Errorf("failed to unmarshal result: %w", err)
	}

	if len(daoResultList) > 1 {
		return *new(D), errors.New("received more than 1 result, expected 1")
	}

	if len(daoResultList) == 0 {
		return *new(D), nil // Return empty object, not an error
	}

	return daoResultList[0], nil
}

// Delete removes an object matching the hash and sort keys.
func Delete(ctx context.Context, repo repository.Repository, hashKey, sortKey string) error {
	config := repo.Config().Values()
	dbSession := repo.Session().(*DynamoSession).session

	_, err := dbSession.DeleteItem(&awsDynamoDB.DeleteItemInput{
		TableName: aws.String(config["table"]),
		Key: map[string]*awsDynamoDB.AttributeValue{
			config["hashKeyName"]: {S: aws.String(hashKey)},
			config["sortKeyName"]: {S: aws.String(sortKey)},
		},
	})

	return err
}