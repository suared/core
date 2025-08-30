package firestore

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/suared/core/v2/repository"
	"github.com/suared/core/v2/security"
)

// ModelMapping defines the mapping for the simple case where the model is stored directly
// as a sub-field of the BaseDAO.
type ModelMapping[M any] struct {
	Collection string
	DocIDFunc  func(m M) string
}

// DaoMapping defines the mapping for the advanced case where a custom DAO is used.
type DaoMapping[M any, D any] struct {
	Collection    string
	DocIDFunc     func(dao D) string
	FromModelFunc func(ctx context.Context, model M) D
	ToModelFunc   func(dao D) M
}

// Repository is the generic Firestore repository that handles data access.
type Repository[M ModelObject] struct {
	config     repository.Config
	session    repository.Session
	Collection string

	// Internal fields that define the behavior for simple vs. advanced paths.
	toDAO       func(ctx context.Context, model M) (any, error)
	fromDAO     func(dao any) (M, error)
	getDocID    func(dao any) string
	getTemplate func(ctx context.Context, model M) any
	selectOneFn func(ctx context.Context, repo repository.Repository, collection string, docID string) (any, error)
	selectFn    func(ctx context.Context, repo repository.Repository, collection string, strFilterVals ...string) ([]any, error)
}

// Config returns the repository's configuration.
func (r *Repository[M]) Config() repository.Config {
	return r.config
}

// SetSession sets the repository's session.
func (r *Repository[M]) SetSession(session repository.Session) {
	r.session = session
}

// Session returns the repository's session.
func (r *Repository[M]) Session() repository.Session {
	return r.session
}

// NewRepository creates a repository for the simple case where the model is stored directly.
func NewRepository[M ModelObject](config repository.Config, mapping ModelMapping[M]) (*Repository[M], error) {
	repo := &Repository[M]{
		config:     config,
		Collection: mapping.Collection,
	}

	// For the simple path, the DAO is the internal BaseDAO[M].
	repo.toDAO = func(ctx context.Context, model M) (any, error) {
		auth := security.GetAuth(ctx)
		user := auth.GetUser()
		now := time.Now().UTC()

		// Check if it's an insert or update by looking at the existing doc.
		// This is a simplification; a real implementation might need a dedicated Get before Set.
		docID := mapping.DocIDFunc(model)
		existing, err := SelectOne[BaseDAO[M]](ctx, repo, repo.Collection, docID)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("checking for existing document failed: %w", err)
		}

		var createdBy = user
		var createdDt = now
		// If the document exists, preserve its creation info.
		if err == nil && !reflect.ValueOf(existing).IsZero() {
			createdBy = existing.CreatedBy
			createdDt = existing.CreatedDt
		}

		return BaseDAO[M]{
			Model:     model,
			CreatedBy: createdBy,
			UpdatedBy: user,
			CreatedDt: createdDt,
			UpdatedDt: now,
		}, nil
	}

	repo.fromDAO = func(dao any) (M, error) {
		if d, ok := dao.(BaseDAO[M]); ok {
			return d.Model, nil
		}
		if d, ok := dao.(*BaseDAO[M]); ok {
			return d.Model, nil
		}
		return *new(M), fmt.Errorf("unexpected DAO type: %T", dao)
	}

	repo.getDocID = func(dao any) string {
		return mapping.DocIDFunc(dao.(BaseDAO[M]).Model)
	}

	repo.getTemplate = func(ctx context.Context, model M) any {
		return BaseDAO[M]{Model: model}
	}

	repo.selectOneFn = func(ctx context.Context, repo repository.Repository, collection string, docID string) (any, error) {
		return SelectOne[BaseDAO[M]](ctx, repo, collection, docID)
	}

	repo.selectFn = func(ctx context.Context, repo repository.Repository, collection string, strFilterVals ...string) ([]any, error) {
		daos, err := Select[BaseDAO[M]](ctx, repo, collection, strFilterVals...)
		if err != nil {
			return nil, err
		}
		anys := make([]any, len(daos))
		for i, d := range daos {
			anys[i] = d
		}
		return anys, nil
	}

	_, err := CreateClient(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return repo, nil
}

// NewRepositoryWithDAO creates a repository for the advanced case with a custom DAO.
func NewRepositoryWithDAO[M ModelObject, D any](config repository.Config, mapping DaoMapping[M, D]) (*Repository[M], error) {
	repo := &Repository[M]{
		config:     config,
		Collection: mapping.Collection,
	}

	// For the advanced path, the DAO is the user-provided type D.
	repo.toDAO = func(ctx context.Context, model M) (any, error) {
		return mapping.FromModelFunc(ctx, model), nil
	}

	repo.fromDAO = func(dao any) (M, error) {
		if d, ok := dao.(D); ok {
			return mapping.ToModelFunc(d), nil
		}
		if d, ok := dao.(*D); ok {
			return mapping.ToModelFunc(*d), nil
		}
		return *new(M), fmt.Errorf("unexpected DAO type: %T", dao)
	}

	repo.getDocID = func(dao any) string {
		return mapping.DocIDFunc(dao.(D))
	}

	repo.getTemplate = func(ctx context.Context, model M) any {
		// For selects, we need a DAO instance to infer the type, but the model is the input.
		// We create a temporary DAO from the model.
		return mapping.FromModelFunc(ctx, model)
	}

	repo.selectOneFn = func(ctx context.Context, repo repository.Repository, collection string, docID string) (any, error) {
		return SelectOne[D](ctx, repo, collection, docID)
	}

	repo.selectFn = func(ctx context.Context, repo repository.Repository, collection string, strFilterVals ...string) ([]any, error) {
		daos, err := Select[D](ctx, repo, collection, strFilterVals...)
		if err != nil {
			return nil, err
		}
		anys := make([]any, len(daos))
		for i, d := range daos {
			anys[i] = d
		}
		return anys, nil
	}

	_, err := CreateClient(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return repo, nil
}

// Insert adds a new document.
func (r *Repository[M]) Insert(ctx context.Context, model M) error {
	dao, err := r.toDAO(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to convert model to DAO: %w", err)
	}

	if err := ValidAction(ctx, "insert", dao); err != nil {
		return err
	}

	docID := r.getDocID(dao)
	return InsertOrUpdate(ctx, r, r.Collection, docID, dao)
}

// Update modifies an existing document.
func (r *Repository[M]) Update(ctx context.Context, model M) error {
	dao, err := r.toDAO(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to convert model to DAO: %w", err)
	}

	if err := ValidAction(ctx, "update", dao); err != nil {
		return err
	}

	docID := r.getDocID(dao)
	return InsertOrUpdate(ctx, r, r.Collection, docID, dao)
}

// Delete removes a document.
func (r *Repository[M]) Delete(ctx context.Context, template M) error {
	dao := r.getTemplate(ctx, template)
	docID := r.getDocID(dao)

	// We need the full DAO for the security check, so we have to get it first.
	fullDAO, err := r.selectOneFn(ctx, r, r.Collection, docID)
	if err != nil {
		return fmt.Errorf("failed to get document for deletion check: %w", err)
	}

	if err := ValidAction(ctx, "delete", fullDAO); err != nil {
		return err
	}

	return Delete(ctx, r, r.Collection, docID)
}

// SelectOne retrieves a single document.
func (r *Repository[M]) SelectOne(ctx context.Context, template M) (M, error) {
	dao := r.getTemplate(ctx, template)
	docID := r.getDocID(dao)

	resultDAO, err := r.selectOneFn(ctx, r, r.Collection, docID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return *new(M), nil // Return empty model if not found, not an error
		}
		return *new(M), err
	}

	if err := ValidAction(ctx, "selectOne", resultDAO); err != nil {
		return *new(M), err
	}

	return r.fromDAO(resultDAO)
}

// Select retrieves multiple documents based on a filter.
func (r *Repository[M]) Select(ctx context.Context, template M, strFilterVals ...string) ([]M, error) {
	resultsDAO, err := r.selectFn(ctx, r, r.Collection, strFilterVals...)
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
		model, err := r.fromDAO(dao)
		if err != nil {
			return nil, err
		}
		outputList = append(outputList, model)
	}

	return outputList, nil
}