package apperix

import (
	"fmt"
	"database/sql"
	"github.com/hashicorp/golang-lru"
)

type ownerProvider struct {
	db *sql.DB
	cache *lru.ARCCache
}

/*
	initialize initializes the owner provider.
	Must be run before usage.
*/
func (provider *ownerProvider) initialize(
	db *sql.DB,
	cacheSize int,
) (
	err error,
) {
	cache, err := lru.NewARC(cacheSize)
	if err != nil {
		return fmt.Errorf("Could not initialize cache: %s", err)
	}
	provider.db = db
	provider.cache = cache
	return nil
}

/*
	GetOwnerOf returns the owner of the given resource
*/
func (provider *ownerProvider) GetOwnerOf(
	resourceId ResourceIdentifier,
) (
	ownerId Identifier,
	err error,
) {
	serializedResId := resourceId.Serialize()

	//cache lookup
	fromCache, exists := provider.cache.Get(serializedResId)
	if exists {
		return fromCache.(Identifier), nil
	}

	//gather from database
	statement, err := provider.db.Prepare(`
		SELECT owner_id FROM resources
		WHERE str_id = ?;
	`)
	defer statement.Close()
	if err != nil {
		return ownerId, DatabaseFailureError {
			message: fmt.Sprintf("Coult not prepare statement: %s", err),
		}
	}
	rows, err := statement.Query(serializedResId)
	defer rows.Close()
	if err != nil {
		return ownerId, DatabaseFailureError {
			message: fmt.Sprintf("Coult not query database: %s", err),
		}
	}
	var ownerIdAsString *string
	rowCount := 0
	for rows.Next() {
		err = rows.Scan(
			&ownerIdAsString,
		)
		if err != nil {
			return ownerId, DatabaseFailureError {
				message: fmt.Sprintf("Coult not scan row: %s", err),
			}
		}
		if ownerIdAsString != nil {
			ownerId.FromString(*ownerIdAsString)
			rowCount++
		}
	}
	if rowCount < 1 {
		return ownerId, NotFoundError {
			message: fmt.Sprintf(
				"Not owner for resource '%s'",
				resourceId.String(),
			),
		}
	}

	//fill cache
	provider.cache.Add(serializedResId, ownerId)

	return ownerId, nil
}