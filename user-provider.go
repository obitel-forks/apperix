package apperix

import (
	"fmt"
	"database/sql"
	"github.com/hashicorp/golang-lru"
)

/*
	The UserAccount type represents bundled information
	about a user registered in the apperix service.
*/
type UserAccount struct {
	Identifier Identifier
	Username string
	Password string
}

type userProvider struct {
	db *sql.DB
	cache *lru.ARCCache
}

/*
	initialize initializes the user provider.
	Must be run before usage.
*/
func (provider *userProvider) initialize(
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
	FindUserById returns a user account identified by the given identifier.
	An error will be returned in case no user was found.
	Tries to return from cache, fills cache on miss.
*/
func (provider *userProvider) FindUserById(
	identifier Identifier,
) (
	account UserAccount,
	err error,
) {
	//cache lookup
	fromCache, exists := provider.cache.Get(identifier)
	if exists {
		return fromCache.(UserAccount), nil
	}

	//gather from database
	statement, err := provider.db.Prepare(`
		SELECT * FROM users
		WHERE identifier = ?
	`)
	defer statement.Close()
	if err != nil {
		return account, DatabaseFailureError {
			message: fmt.Sprintf("Coult not prepare statement: %s", err),
		}
	}
	rows, err := statement.Query(identifier.String())
	defer rows.Close()
	if err != nil {
		return account, DatabaseFailureError {
			message: fmt.Sprintf("Coult not query database: %s", err),
		}
	}
	rowCount := 0
	for rows.Next() {
		var id string
		err = rows.Scan(
			&id,
			&account.Username,
			&account.Password,
		)
		account.Identifier.FromString(id)
		if err != nil {
			return account, DatabaseFailureError {
				message: fmt.Sprintf("Coult not scan row: %s", err),
			}
		}
		rowCount++
	}
	if rowCount < 1 {
		return account, NotFoundError {
			message: fmt.Sprintf(
				"User identified by identifier '%s' not found",
				identifier.String(),
			),
		}
	}

	//update cache
	provider.cache.Add(identifier, account)

	return account, nil
}

/*
	FindUserByUsername returns a user account identified by the given username.
	An error will be returned in case no user was found.
	Tries to return from cache, fills cache on miss.
*/
func (provider *userProvider) FindUserByUsername(
	username string,
) (
	account UserAccount,
	err error,
) {
	//cache lookup
	fromCache, exists := provider.cache.Get(username)
	if exists {
		return fromCache.(UserAccount), nil
	}

	//gather from database
	statement, err := provider.db.Prepare(`
		SELECT * FROM users
		WHERE username = ?
	`)
	defer statement.Close()
	if err != nil {
		return account, fmt.Errorf("Coult not prepare statement: %s", err)
	}
	rows, err := statement.Query(username)
	defer rows.Close()
	if err != nil {
		return account, fmt.Errorf("Coult not query database: %s", err)
	}
	numberOfUsers := 0
	for rows.Next() {
		var id string
		err = rows.Scan(
			&id,
			&account.Username,
			&account.Password,
		)
		account.Identifier.FromString(id)
		if err != nil {
			return account, fmt.Errorf("Coult not scan row: %s", err)
		}
		numberOfUsers++
	}
	if numberOfUsers < 1 {
		return account, fmt.Errorf(
			"User identified by username '%s' not found",
			username,
		)
	}

	//update cache
	provider.cache.Add(username, account)

	return account, nil
}