package apperix

import (
	"fmt"
	"bytes"
	"strings"
	"database/sql"
	"github.com/satori/go.uuid"
)

type transaction struct {
	database *sql.DB
	identifier []byte
}

func (txn *transaction) Begin() {
	txn.identifier = []byte(strings.Replace(uuid.NewV4().String(), "-", "", -1))
	var query bytes.Buffer
	query.WriteString("SAVEPOINT \"")
	query.Write(txn.identifier)
	query.WriteRune('"')
	_, err := txn.database.Exec(query.String())
	if err != nil {
		panic(fmt.Errorf(
			"Could not begin transaction ('%s'): %s",
			string(txn.identifier),
			err,
		))
	}
}

func (txn *transaction) Commit() {
	var query bytes.Buffer
	query.WriteString("RELEASE SAVEPOINT \"")
	query.Write(txn.identifier)
	query.WriteRune('"')
	_, err := txn.database.Exec(query.String())
	if err != nil {
		panic(fmt.Errorf(
			"Could not commit transaction ('%s'): %s",
			string(txn.identifier),
			err,
		))
	}
}

func (txn *transaction) Rollback() {
	var query bytes.Buffer
	query.WriteString("ROLLBACK TO SAVEPOINT \"")
	query.Write(txn.identifier)
	query.WriteRune('"')
	_, err := txn.database.Exec(query.String())
	if err != nil {
		panic(fmt.Errorf(
			"Could not rollback transaction ('%s'): %s",
			string(txn.identifier),
			err,
		))
	}
}