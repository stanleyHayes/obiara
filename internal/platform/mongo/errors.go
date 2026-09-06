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
// there. namespaceNotFoundCode is its reply when the collection holding that
// index does not exist either, which is what a genuinely empty database
// answers — and which is the case this predicate always claimed to cover.
const (
	indexNotFoundCode     = 27
	namespaceNotFoundCode = 26
)

// IsIndexNotFound reports whether err is a missing-index error from a drop.
// Repositories use it to make index removal idempotent: a schema change that
// retires an index must not fail the boot of an instance that has already
// applied it, or of a fresh database that never had it.
func IsIndexNotFound(err error) bool {
	var commandError mongo.CommandError
	if errors.As(err, &commandError) {
		// A missing collection means the index is missing too, and it is the
		// answer a fresh database gives. Without this the first boot against
		// an empty database fails on a drop that was always meant to be
		// tolerated — which is to say the product could not be deployed to a
		// new environment at all (agent_plan.md §80).
		return commandError.Code == indexNotFoundCode ||
			commandError.Code == namespaceNotFoundCode ||
			commandError.Name == "IndexNotFound" ||
			commandError.Name == "NamespaceNotFound"
	}
	return false
}
