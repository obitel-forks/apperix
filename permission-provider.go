package apperix

import (
	"fmt"
	"bytes"
	"database/sql"
	"github.com/hashicorp/golang-lru"
)

type permissionProvider struct {
	db *sql.DB
	cache *lru.ARCCache
}

/*
	initialize initializes the permissions provider.
	Must be run before usage.
*/
func (provider *permissionProvider) initialize(
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

func (provider *permissionProvider) GetPermissionsFor(
	resourceId ResourceIdentifier,
	user string,
) (
	permissions Permissions,
	err error,
) {
	var cacheKeyBuf bytes.Buffer
	cacheKeyBuf.WriteString(user)
	cacheKeyBuf.WriteRune(':')
	cacheKeyBuf.WriteString(resourceId.Serialize())
	cacheKey := cacheKeyBuf.String()

	//cache lookup
	fromCache, exists := provider.cache.Get(cacheKey)
	if exists {
		return fromCache.(Permissions), nil
	}

	//gather from database
	statement, err := provider.db.Prepare(`
		SELECT permissions FROM resource_permissions
		WHERE resource_id = (SELECT id FROM resources WHERE str_id = ?)
		AND user_id = ?;
	`)
	defer statement.Close()
	if err != nil {
		return permissions, DatabaseFailureError {
			message: fmt.Sprintf("Coult not prepare statement: %s", err),
		}
	}
	rows, err := statement.Query(resourceId.Serialize(), user)
	defer rows.Close()
	if err != nil {
		return permissions, DatabaseFailureError {
			message: fmt.Sprintf("Coult not query database: %s", err),
		}
	}
	var encodedPermissions uint32
	rowCount := 0
	for rows.Next() {
		err = rows.Scan(
			&encodedPermissions,
		)
		permissions.Deserialize(encodedPermissions)
		if err != nil {
			return permissions, DatabaseFailureError {
				message: fmt.Sprintf("Coult not scan row: %s", err),
			}
		}
		rowCount++
	}
	if rowCount < 1 {
		return permissions, NotFoundError {
			message: fmt.Sprintf("No result for user '%s' on resource '%s'",
				user,
				resourceId.String(),
			),
		}
	}

	//fill cache
	provider.cache.Add(cacheKey, permissions)

	return permissions, nil
}