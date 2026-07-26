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
