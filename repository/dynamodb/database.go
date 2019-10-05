package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awsDynamoDB "github.com/aws/aws-sdk-go/service/dynamodb"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	_ "github.com/suared/core/infra"
	"github.com/suared/core/repository"
	"github.com/suared/core/security"
)

//DAO marker interface - Dynamo specific for now, might make sense at some point to find generic names to work across
type DAO interface {
	HashKey() string
	SortKey() string
	User() string
	//New - Enables a copy of the specific struct to support return values
	New() DAO
	//Refresh - Updates the Hash/SortKey based on current values
	Refresh()
}

// DaoAudit - INTERNAL ONLY.  Groups Audit data, exported to support default JSON conversion.
type DaoAudit struct {
	//These are audit columns
	CreatedBy string
	UpdatedBy string
	//TODO: confirm default go type includes timezone approach or update to alt that does
	CreatedDt time.Time
	UpdatedDt time.Time
}

//DynamoSession - Implementation of DynamoSession
type DynamoSession struct {
	session *awsDynamoDB.DynamoDB
}

//Session - Return this session/ implement the Session interface
func (s *DynamoSession) Session() repository.Session {
	return s
}

func newDynamoSession(awsDynamoDBSession *awsDynamoDB.DynamoDB) *DynamoSession {
	dynSession := new(DynamoSession)
	dynSession.session = awsDynamoDBSession
	return dynSession
}

//ValidAction - Default Security check
func ValidAction(ctx context.Context, action string, dao DAO) error {
	//Get authentication object
	auth := security.GetAuth(ctx)
	authUser := auth.GetUser()
	daoUser := dao.User()

	//If match -- we are good
	if authUser == daoUser {
		return nil
	}

	//Check if Admin
	if auth.IsAdmin() {
		return nil
	}

	// Security checks did not succeed
	return fmt.Errorf("Security: User: %v does not have access to %v for %v", authUser, action, daoUser)
}

//CreateTable - will create a dynamodb table if it doesn't exist or return the existing table interface if it does
func CreateTable(repo repository.Repository) (repository.Repository, error) {
	// First Get & Check Parameters
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

	rcu := config["rcu"]
	RCU, err := strconv.ParseInt(rcu, 10, 64)

	if err != nil {
		return nil, errors.New("rcu is not a number")
	}

	wcu := config["wcu"]
	WCU, err := strconv.ParseInt(wcu, 10, 64)

	if err != nil {
		return nil, errors.New("wcu is not a number")
	}

	env := config["env"]
	if env == "" || env == "dev" || env == "development" {
		env = "dev"
	}

	//Save Enviroment for AWS for future calls
	// Library Reference: https://github.com/aws/aws-sdk-go
	// Endpoint only required for local dev, default is empty string so will defer to env instead
	// LogLevel default is 0, will not set that for now assuming errors will be descriptive enough for local dev
	// Defer credentials to the default AWS search chain
	awsConfig := aws.Config{Region: aws.String(region), Endpoint: aws.String(endpoint)}
	//
	//     // Specify profile for config and region for requests
	//     sess := session.Must(session.NewSessionWithOptions(session.Options{
	//          Config: aws.Config{Region: aws.String("us-east-1")},
	//          Profile: "profile_name",
	//     }))
	//
	awsDynamoSession := awsSession.Must(awsSession.NewSessionWithOptions(awsSession.Options{Config: awsConfig}))
	awsDynamoService := awsDynamoDB.New(awsDynamoSession)

	repo.SetSession(newDynamoSession(awsDynamoService))

	if env != "dev" {
		//For non-dev, expect infra scripts have created infra
		return repo, nil
	}

	// Create table if appropriate (note: shortcut above for dev)
	req := &awsDynamoDB.DescribeTableInput{TableName: aws.String(table)}
	_, err = awsDynamoService.DescribeTable(req)

	// An error here signifies that the table does not yet exist
	if err != nil {
		log.Printf("Generating datamodel for enviroment: %v, endpoint: %v, with tablename: %v...",
			region,
			endpoint,
			table)
		readcapacity := RCU
		writecapacity := WCU

		//Table does not exist in this environment, create it...
		awsTableCreateParams := &awsDynamoDB.CreateTableInput{
			AttributeDefinitions: []*awsDynamoDB.AttributeDefinition{
				{
					AttributeName: aws.String(hashKeyName),
					AttributeType: aws.String("S"),
				},
				{
					AttributeName: aws.String(sortKeyName),
					AttributeType: aws.String("S"),
				},
			},
			KeySchema: []*awsDynamoDB.KeySchemaElement{
				{
					AttributeName: aws.String(hashKeyName),
					KeyType:       aws.String("HASH"),
				},
				{
					AttributeName: aws.String(sortKeyName),
					KeyType:       aws.String("RANGE"),
				},
			},
			ProvisionedThroughput: &awsDynamoDB.ProvisionedThroughput{
				ReadCapacityUnits:  aws.Int64(readcapacity),
				WriteCapacityUnits: aws.Int64(writecapacity),
			},
			TableName: aws.String(table),
		}

		_, err = awsDynamoService.CreateTable(awsTableCreateParams)

		if err != nil {
			panic(err)
		}
	}

	//After initial setup the table should be found, if not something is seriously wrong
	_, err = awsDynamoService.DescribeTable(req)
	if err != nil {
		panic("Unable to setup dynamo process table on init")
	}

	return repo, nil
}

//InsertOrUpdate - Generic method to insert or update a dynamo table
func InsertOrUpdate(ctx context.Context, repo repository.Repository, dao DAO) error {
	dao.Refresh()
	//  Save to the Database
	awsDAO, err := dynamodbattribute.MarshalMap(dao)

	input := &awsDynamoDB.PutItemInput{
		Item:      awsDAO,
		TableName: aws.String(repo.Config().Values()["table"]),
	}

	dbSession := repo.Session().(*DynamoSession).session
	_, err = dbSession.PutItem(input)

	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

//Select - Returns a list of DAO objects matching the template hashKey.  Validation is expected to be done by the caller
func Select(ctx context.Context, repo repository.Repository, templ DAO) ([]DAO, error) {
	templ.Refresh()

	hashKey := templ.HashKey()

	input := &awsDynamoDB.QueryInput{
		ExpressionAttributeValues: map[string]*awsDynamoDB.AttributeValue{
			":column": {
				S: aws.String(hashKey),
			},
		},
		//"UserActive = :column"
		KeyConditionExpression: aws.String(repo.Config().Values()["hashKeyName"] + " = :column"),
		// not needed here, this is to set specific columns to return, example for ony GUID:	ProjectionExpression:   aws.String("GUID"),
		TableName: aws.String(repo.Config().Values()["table"]),
	}

	dbSession := repo.Session().(*DynamoSession).session
	result, err := dbSession.Query(input)
	if err != nil {
		return nil, err
	}

	//Note: validation is deferred to the caller in case the default validation needs to be updated for a specific case
	var daoResultList []DAO

	for i := range result.Items {

		resultDAO := templ.New()

		err = dynamodbattribute.UnmarshalMap(result.Items[i], resultDAO)
		if err != nil {
			//	log.Printf("Error unmarshalling dynamo data: %v", err)
			return nil, err
		}

		resultDAO.Refresh()

		daoResultList = append(daoResultList, resultDAO)

		//log.Printf("output list: %v", outputList)
	}

	return daoResultList, nil

}

//SelectOne - Returns a DAO object matching the template hashKey and sortKey.  Validation is expected to be done by the caller
func SelectOne(ctx context.Context, repo repository.Repository, templ DAO) (DAO, error) {
	templ.Refresh()

	hashKey := templ.HashKey()
	sortKey := templ.SortKey()

	input := &awsDynamoDB.QueryInput{
		ExpressionAttributeValues: map[string]*awsDynamoDB.AttributeValue{
			":thehash": {
				S: aws.String(hashKey),
			},
			":thesort": {
				S: aws.String(sortKey),
			},
		},
		//"UserActive = :column"
		KeyConditionExpression: aws.String(repo.Config().Values()["hashKeyName"] + " = :thehash AND " +
			repo.Config().Values()["sortKeyName"] + " = :thesort"),
		// not needed here, this is to set specific columns to return, example for ony GUID:	ProjectionExpression:   aws.String("GUID"),
		TableName: aws.String(repo.Config().Values()["table"]),
	}

	dbSession := repo.Session().(*DynamoSession).session
	result, err := dbSession.Query(input)
	if err != nil {
		return nil, err
	}

	//Note: validation is deferred to the caller in case the default validation needs to be updated for a specific case
	var daoResultList []DAO

	for i := range result.Items {

		resultDAO := templ.New()

		err = dynamodbattribute.UnmarshalMap(result.Items[i], resultDAO)
		if err != nil {
			//	log.Printf("Error unmarshalling dynamo data: %v", err)
			return nil, err
		}

		resultDAO.Refresh()
		daoResultList = append(daoResultList, resultDAO)

		//log.Printf("output list: %v", outputList)
	}

	if len(daoResultList) > 1 {
		return nil, errors.New("Received more than 1 result, expected 1.  Fix the base query")
	}

	return daoResultList[0], nil

}

//Delete - Removes a DAO object matching the template hashKey and sortKey.  Validation is expected to be done by the caller.  May consider validation of deletion in te future, intentionally not there now
func Delete(ctx context.Context, repo repository.Repository, templ DAO) error {
	templ.Refresh()

	hashKey := templ.HashKey()
	sortKey := templ.SortKey()

	dbSession := repo.Session().(*DynamoSession).session

	_, err := dbSession.DeleteItem(&awsDynamoDB.DeleteItemInput{
		TableName: aws.String(repo.Config().Values()["table"]),
		Key: map[string]*awsDynamoDB.AttributeValue{
			repo.Config().Values()["hashKeyName"]: {
				S: aws.String(hashKey),
			},
			repo.Config().Values()["sortKeyName"]: {
				S: aws.String(sortKey),
			},
		},
	})

	if err != nil {
		return err
	}

	return nil

}
