package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/sorrtory/freebooru/internal/collection"
	querysearch "github.com/sorrtory/freebooru/internal/search"
)

// DefaultSearchLimit is used when a frontend does not specify a limit.
const DefaultSearchLimit int64 = 100

// FileSearchRequest contains frontend-preserved query terms and pagination.
// A nil Limit selects DefaultSearchLimit; zero is an explicit empty page.
type FileSearchRequest struct {
	Collection string
	Terms      []string
	Limit      *int64
	Offset     int64
}

// Search parses, resolves, and executes one validated collection query.
func (c *Core) Search(
	ctx context.Context,
	request FileSearchRequest,
) (files []collection.FileRecord, err error) {
	collectionName, references, _, err := c.tagMutationCollection(request.Collection)
	if err != nil {
		return nil, err
	}
	query, err := querysearch.Parse(request.Terms)
	if err != nil {
		return nil, err
	}
	resolved, err := querysearch.Resolve(query, c.catalog, references)
	if err != nil {
		return nil, err
	}
	limit := DefaultSearchLimit
	if request.Limit != nil {
		limit = *request.Limit
	}
	repositoryRequest, err := collectionSearchRequest(resolved, limit, request.Offset)
	if err != nil {
		return nil, err
	}
	database, err := c.OpenCollection(ctx, collectionName)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err, closeNamedCollection(database, collectionName))
	}()
	files, err = database.Search(ctx, repositoryRequest)
	if err != nil {
		return nil, fmt.Errorf("search collection %q: %w", collectionName, err)
	}
	return files, nil
}

func collectionSearchRequest(
	query querysearch.ResolvedQuery,
	limit int64,
	offset int64,
) (collection.SearchRequest, error) {
	if limit < 0 {
		return collection.SearchRequest{}, fmt.Errorf("search limit must be non-negative")
	}
	if offset < 0 {
		return collection.SearchRequest{}, fmt.Errorf("search offset must be non-negative")
	}
	request := collection.SearchRequest{
		Terms:  make([]collection.SearchTerm, 0, len(query.Terms)),
		Limit:  limit,
		Offset: offset,
	}
	for _, term := range query.Terms {
		operator, err := collectionSearchOperator(term.Operator)
		if err != nil {
			return collection.SearchRequest{}, err
		}
		request.Terms = append(request.Terms, collection.SearchTerm{
			Tag:      term.Tag,
			Type:     string(term.Type),
			Operator: operator,
			Value:    term.Value,
		})
	}
	return request, nil
}

func collectionSearchOperator(operator querysearch.Operator) (collection.SearchOperator, error) {
	switch operator {
	case querysearch.Present:
		return collection.SearchPresent, nil
	case querysearch.Absent:
		return collection.SearchAbsent, nil
	case querysearch.Equal:
		return collection.SearchEqual, nil
	case querysearch.Less:
		return collection.SearchLess, nil
	case querysearch.LessEqual:
		return collection.SearchLessEqual, nil
	case querysearch.Greater:
		return collection.SearchGreater, nil
	case querysearch.GreaterEqual:
		return collection.SearchGreaterEqual, nil
	default:
		return "", fmt.Errorf("search operator %q is unsupported", operator)
	}
}
