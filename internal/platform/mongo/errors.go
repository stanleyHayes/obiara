package mongo

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

// duplicateKeyCode is the MongoDB server code for unique-index violations.
const duplicateKeyCode = 11000

// IsDuplicateKey reports whether err is a unique-index violation. Module
// repositories use it to translate driver errors into domain-meaningful
// results (e.g. idempotent re-registration, duplicate room rejection)
// instead of leaking driver types upward (agent_plan.md §7.2).
func IsDuplicateKey(err error) bool {
	var writeException mongo.WriteException
	if errors.As(err, &writeException) {
		for _, writeError := range writeException.WriteErrors {
			if writeError.Code == duplicateKeyCode {
				return true
			}
		}
	}
	var commandError mongo.CommandError
	if errors.As(err, &commandError) && commandError.Code == duplicateKeyCode {
		return true
	}
	return false
}

// indexNotFoundCode is the server's reply to dropping an index that is not
// there.
const indexNotFoundCode = 27

// IsIndexNotFound reports whether err is a missing-index error from a drop.
// Repositories use it to make index removal idempotent: a schema change that
// retires an index must not fail the boot of an instance that has already
// applied it, or of a fresh database that never had it.
func IsIndexNotFound(err error) bool {
	var commandError mongo.CommandError
	if errors.As(err, &commandError) {
		return commandError.Code == indexNotFoundCode ||
			commandError.Name == "IndexNotFound"
	}
	return false
}
