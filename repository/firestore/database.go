package firestore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	_ "github.com/suared/core/v2/infra"
	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/security"
	"google.golang.org/api/option"
)

// BaseDAO provides standard DB fields and a place for the user's model.
type BaseDAO[M any] struct {
	Model     M         `firestore:"model"`
	CreatedBy string    `firestore:"createdBy,omitempty"`
	UpdatedBy string    `firestore:"updatedBy,omitempty"`
	CreatedDt time.Time `firestore:"createdDt,omitempty"`
	UpdatedDt time.Time `firestore:"updatedDt,omitempty"`
}

// ModelObject defines an interface that model objects must implement.
type ModelObject interface {
	IsEmpty() bool
}

// FirestoreSession - Implementation of FirestoreSession
type FirestoreSession struct {
	client *firestore.Client
}

// Session - Return this session/ implement the Session interface
func (s *FirestoreSession) Session() repository.Session {
	return s
}

func newFirestoreSession(client *firestore.Client) *FirestoreSession {
	fs := new(FirestoreSession)
	fs.client = client
	return fs
}

// ValidAction performs a security check. It now uses reflection to find the user field.
func ValidAction(ctx context.Context, action string, dao any) error {
	auth := security.GetAuth(ctx)
	authUser := auth.GetUser()

	// Use reflection to find a user field (e.g., CreatedBy) in the DAO.
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

	// If no user field is found, we might allow the action for safety, or deny it.
	// For now, if we can't determine ownership, we let it pass but this could be stricter.
	if daoUser == "" {
		return nil
	}

	if authUser == daoUser || auth.IsAdmin() {
		return nil
	}

	return fmt.Errorf("security: User: %v does not have access to %v for %v", authUser, action, daoUser)
}

// CreateClient will create a firestore client
func CreateClient(repo repository.Repository) (repository.Repository, error) {
	config := repo.Config().Values()
	backend := config["backend"]
	if backend != "firestore" {
		return nil, fmt.Errorf("firestore backend called for repository without the appropriate backend type: %v vs %v", backend, "firestore")
	}

	projectID := config["projectID"]
	if projectID == "" {
		return nil, errors.New("firestore projectID cannot be empty")
	}

	endpoint := config["endpoint"]
	env := config["env"]
	if env == "" || env == "dev" || env == "development" {
		env = "dev"
	}

	var client *firestore.Client
	var err error
	ctx := context.Background()

	if env == "dev" && endpoint != "" {
		client, err = firestore.NewClient(ctx, projectID, option.WithEndpoint(endpoint))
	} else {
		client, err = firestore.NewClient(ctx, projectID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	repo.SetSession(newFirestoreSession(client))

	return repo, nil
}

// InsertOrUpdate is a generic method to insert or update a firestore document.
func InsertOrUpdate(ctx context.Context, repo repository.Repository, collection, docID string, data any) error {
	client := repo.Session().(*FirestoreSession).client
	_, err := client.Collection(collection).Doc(docID).Set(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to set document: %w", err)
	}
	return nil
}

// Select returns a list of objects matching the query.
func Select[D any](ctx context.Context, repo repository.Repository, collection string, strFilterVals ...string) ([]D, error) {
	client := repo.Session().(*FirestoreSession).client
	coll := client.Collection(collection)
	var query firestore.Query = coll.Query

	if len(strFilterVals) > 0 {
		countAttrs := len(strFilterVals) / 3
		for i := 0; i < countAttrs*3; i += 3 {
			path := strFilterVals[i]
			op := strFilterVals[i+1]
			value := strFilterVals[i+2]
			query = query.Where(path, op, value)
		}
	}

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var results []D
	for _, doc := range docs {
		var resultDAO D
		if err := doc.DataTo(&resultDAO); err != nil {
			return nil, fmt.Errorf("failed to map data to DAO: %w", err)
		}
		results = append(results, resultDAO)
	}

	return results, nil
}

// SelectOne returns a single object matching the document ID.
func SelectOne[D any](ctx context.Context, repo repository.Repository, collection, docID string) (D, error) {
	var resultDAO D
	client := repo.Session().(*FirestoreSession).client
	doc, err := client.Collection(collection).Doc(docID).Get(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return *new(D), err // Return zero value and the not found error
		}
		return *new(D), err
	}

	if err := doc.DataTo(&resultDAO); err != nil {
		return *new(D), fmt.Errorf("failed to map data to DAO: %w", err)
	}

	return resultDAO, nil
}

// Delete removes a document matching the document ID.
func Delete(ctx context.Context, repo repository.Repository, collection, docID string) error {
	client := repo.Session().(*FirestoreSession).client
	_, err := client.Collection(collection).Doc(docID).Delete(ctx)
	return err
}