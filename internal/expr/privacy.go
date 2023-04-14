package expr

import (
	"github.com/laytan/php-parser/pkg/ast"
	"github.com/laytan/phpls/pkg/phprivacy"
)

// Context about the iteration for determining privacy.
type iteration struct {
	// Is this iteration the first?
	first bool
	// Is this iteration the first ir.ClassStmt we see?
	firstClass bool
}

// Determines the privacy to search for based on all the conditions determined
// by PHP.
func determinePrivacy(
	startPrivacy phprivacy.Privacy,
	currKind ast.Type,
	iteration *iteration,
) phprivacy.Privacy {
	actPrivacy := startPrivacy

	// If we are in the class, the first run can check >= private members,
	// the rest only >= protected members.
	if !iteration.first && actPrivacy == phprivacy.PrivacyPrivate {
		actPrivacy = phprivacy.PrivacyProtected
	}

	// If this is a trait, and it is used from the first class,
	// private methods are also accessible.
	if iteration.firstClass && actPrivacy == phprivacy.PrivacyProtected &&
		currKind == ast.TypeStmtTrait {
		actPrivacy = phprivacy.PrivacyPrivate
	}

	return actPrivacy
}
