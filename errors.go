package apperix

/*
	NotFoundError represents error cases where the requested
	object was not found.
*/
type NotFoundError struct {
	message string
}

func (err NotFoundError) Error() string {
	return err.message
}

/*
	DatabaseFailureError represents error cases where
	either the database is unreachable or the database query was faulty.
*/
type DatabaseFailureError struct {
	message string
}

func (err DatabaseFailureError) Error() string {
	return err.message
}