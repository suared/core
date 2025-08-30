package dynamodb

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/security"
)

// DaoMapping defines the mapping between a user model and a custom DAO struct.
// For DynamoDB, a custom DAO is required to define the hash and sort key fields.
type DaoMapping[M any, D any] struct {
	// FromModelFunc converts the user-facing model into the database DAO.
	FromModelFunc func(ctx context.Context, model M) D
	// ToModelFunc converts the database DAO back into the user-facing model.
	ToModelFunc func(dao D) M
	// HashKeyFunc extracts the hash key value from the DAO.
	HashKeyFunc func(dao D) string
	// SortKeyFunc extracts the sort key value from the DAO.
	SortKeyFunc func(dao D) string
}

// Repository is the generic DynamoDB repository.
// It is generic for the model type M and the DAO type D.
type Repository[M ModelObject, D any] struct {
	config  repository.Config
	session repository.Session
	mapping DaoMapping[M, D]
}

// Config returns the repository's configuration.
func (r *Repository[M, D]) Config() repository.Config {
	return r.config
}

// SetSession sets the repository's session.
func (r *Repository[M, D]) SetSession(session repository.Session) {
	r.session = session
}

// Session returns the repository's session.
func (r *Repository[M, D]) Session() repository.Session {
	return r.session
}

// NewRepository creates a new DynamoDB repository with a custom DAO mapping.
func NewRepository[M ModelObject, D any](config repository.Config, mapping DaoMapping[M, D]) (*Repository[M, D], error) {
	repo := &Repository[M, D]{
		config:  config,
		mapping: mapping,
	}

	_, err := CreateTable(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize dynamodb table: %w", err)
	}

	return repo, nil
}

// setAuditFields uses reflection to set audit fields on a DAO if it embeds BaseDAO.
func setAuditFields(ctx context.Context, dao any) {
	auth := security.GetAuth(ctx)
	user := auth.GetUser()
	now := time.Now().UTC()

	v := reflect.ValueOf(dao)
	if v.Kind() != reflect.Ptr {
		return // Must be a pointer to a struct
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return
	}

	// Look for the embedded BaseDAO field.
	baseField := v.FieldByName("BaseDAO")
	if !baseField.IsValid() || baseField.Kind() != reflect.Struct {
		return
	}

	// Check if it's a new record by looking at CreatedDt.
	createdDtField := baseField.FieldByName("CreatedDt")
	if createdDtField.IsValid() && createdDtField.CanSet() {
		if createdDtField.Interface().(time.Time).IsZero() {
			// It's a new record, set CreatedBy and CreatedDt.
			createdByField := baseField.FieldByName("CreatedBy")
			if createdByField.IsValid() && createdByField.CanSet() {
				createdByField.SetString(user)
			}
			createdDtField.Set(reflect.ValueOf(now))
		}
	}

	// Always set the UpdatedBy and UpdatedDt fields.
	updatedByField := baseField.FieldByName("UpdatedBy")
	if updatedByField.IsValid() && updatedByField.CanSet() {
		updatedByField.SetString(user)
	}
	updatedDtField := baseField.FieldByName("UpdatedDt")
	if updatedDtField.IsValid() && updatedDtField.CanSet() {
		updatedDtField.Set(reflect.ValueOf(now))
	}
}

// Insert adds a new item to the table.
func (r *Repository[M, D]) Insert(ctx context.Context, model M) error {
	dao := r.mapping.FromModelFunc(ctx, model)

	// Set audit fields if the DAO has an embedded BaseDAO.
	setAuditFields(ctx, &dao)

	if err := ValidAction(ctx, "insert", dao); err != nil {
		return err
	}

	return InsertOrUpdate(ctx, r, dao)
}

// Update modifies an existing item in the table.
func (r *Repository[M, D]) Update(ctx context.Context, model M) error {
	dao := r.mapping.FromModelFunc(ctx, model)

	// Set audit fields if the DAO has an embedded BaseDAO.
	setAuditFields(ctx, &dao)

	if err := ValidAction(ctx, "update", dao); err != nil {
		return err
	}

	return InsertOrUpdate(ctx, r, dao)
}

// Delete removes an item from the table.
func (r *Repository[M, D]) Delete(ctx context.Context, template M) error {
	dao := r.mapping.FromModelFunc(ctx, template)
	hashKey := r.mapping.HashKeyFunc(dao)
	sortKey := r.mapping.SortKeyFunc(dao)

	// We need the full DAO for the security check.
	fullDAO, err := SelectOne[D](ctx, r, hashKey, sortKey)
	if err != nil {
		return fmt.Errorf("failed to get item for deletion check: %w", err)
	}

	if err := ValidAction(ctx, "delete", fullDAO); err != nil {
		return err
	}

	return Delete(ctx, r, hashKey, sortKey)
}

// SelectOne retrieves a single item.
func (r *Repository[M, D]) SelectOne(ctx context.Context, template M) (M, error) {
	dao := r.mapping.FromModelFunc(ctx, template)
	hashKey := r.mapping.HashKeyFunc(dao)
	sortKey := r.mapping.SortKeyFunc(dao)

	resultDAO, err := SelectOne[D](ctx, r, hashKey, sortKey)
	if err != nil {
		return *new(M), err
	}

	if reflect.ValueOf(resultDAO).IsZero() {
		return *new(M), nil // Not found
	}

	if err := ValidAction(ctx, "selectOne", resultDAO); err != nil {
		return *new(M), err
	}

	return r.mapping.ToModelFunc(resultDAO), nil
}

// Select retrieves multiple items based on a hash key and optional filters.
func (r *Repository[M, D]) Select(ctx context.Context, template M, strFilterVals ...string) ([]M, error) {
	dao := r.mapping.FromModelFunc(ctx, template)
	hashKey := r.mapping.HashKeyFunc(dao)

	resultsDAO, err := Select[D](ctx, r, hashKey, strFilterVals...)
	if err != nil {
		return nil, err
	}

	var outputList []M
	validated := false
	for _, dao := range resultsDAO {
		if !validated {
			if err := ValidAction(ctx, "select", dao); err != nil {
				return nil, err
			}
			validated = true
		}
		outputList = append(outputList, r.mapping.ToModelFunc(dao))
	}

	return outputList, nil
}