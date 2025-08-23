package firestore

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	_ "github.com/suared/core/v2/infra"
	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/security"
	"google.golang.org/api/option"
)

// DAO marker interface - Firestore specific for now
type DAO interface {
	Collection() string
	DocID() string
	User() string
	SetUser(context.Context)
	Setup(userModel interface{})
	GetModel() interface{}
	New() DAO
	Refresh()
	Populate()
}

// DaoAudit - INTERNAL ONLY.  Groups Audit data, exported to support default JSON conversion.
type DaoAudit struct {
	CreatedBy string
	UpdatedBy string
	CreatedDt time.Time
	UpdatedDt time.Time
}

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

// ValidAction - Default Security check, called externally to enable swap outs
func ValidAction(ctx context.Context, action string, dao DAO) error {
	auth := security.GetAuth(ctx)
	authUser := auth.GetUser()
	daoUser := dao.User()

	if authUser == daoUser {
		return nil
	}

	if auth.IsAdmin() {
		return nil
	}

	return fmt.Errorf("Security: User: %v does not have access to %v for %v", authUser, action, daoUser)
}

// CreateClient - will create a firestore client
func CreateClient(repo repository.Repository) (repository.Repository, error) {
	config := repo.Config().Values()
	backend := config["backend"]
	if backend != "firestore" {
		return nil, fmt.Errorf("Firestore backend called for repository without the appropriate backend type: %v vs %v", backend, "firestore")
	}

	projectID := config["projectID"]
	if projectID == "" {
		return nil, errors.New("Firestore projectID cannot be empty")
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
		// Use emulator
		client, err = firestore.NewClient(ctx, projectID, option.WithEndpoint(endpoint))
	} else {
		// Connect to actual Firestore
		client, err = firestore.NewClient(ctx, projectID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %v", err)
	}

	repo.SetSession(newFirestoreSession(client))

	return repo, nil
}

// InsertOrUpdate - Generic method to insert or update a firestore document
func InsertOrUpdate(ctx context.Context, repo repository.Repository, dao DAO) error {
	dao.Refresh()
	client := repo.Session().(*FirestoreSession).client
	_, err := client.Collection(dao.Collection()).Doc(dao.DocID()).Set(ctx, dao)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

// Select - Returns a list of DAO objects matching the query.
func Select(ctx context.Context, repo repository.Repository, templ DAO, strFilterVals ...string) ([]DAO, error) {
	templ.Refresh()
	client := repo.Session().(*FirestoreSession).client
	coll := client.Collection(templ.Collection())
	var query firestore.Query = coll.Query

	if strFilterVals != nil {
		countAttrs := len(strFilterVals) / 3
		for i := 0; i < countAttrs*3; i = i + 3 {
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

	var daoResultList []DAO
	for _, doc := range docs {
		resultDAO := templ.New()
		if err := doc.DataTo(resultDAO); err != nil {
			return nil, err
		}
		resultDAO.Refresh()
		resultDAO.Populate()
		daoResultList = append(daoResultList, resultDAO)
	}

	return daoResultList, nil
}

// SelectOne - Returns a DAO object matching the template docID.
func SelectOne(ctx context.Context, repo repository.Repository, templ DAO) (DAO, error) {
	templ.Refresh()
	client := repo.Session().(*FirestoreSession).client
	doc, err := client.Collection(templ.Collection()).Doc(templ.DocID()).Get(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return templ.New(), nil
		}
		return nil, err
	}

	resultDAO := templ.New()
	if err := doc.DataTo(resultDAO); err != nil {
		return nil, err
	}

	resultDAO.Refresh()
	resultDAO.Populate()
	return resultDAO, nil
}

// Delete - Removes a DAO object matching the template docID.
func Delete(ctx context.Context, repo repository.Repository, templ DAO) error {
	templ.Refresh()
	client := repo.Session().(*FirestoreSession).client
	_, err := client.Collection(templ.Collection()).Doc(templ.DocID()).Delete(ctx)
	return err
}
