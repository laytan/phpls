package phprivacy

import (
	"fmt"
)

type Privacy int

const (
	PrivacyPublic Privacy = iota
	PrivacyProtected
	PrivacyPrivate
)

// If I am calling from the p scope relative to the pb class, is that allowed?
//
// For examples:
//   p = PrivacyPublic,  pb = PrivacyPrivate -> false
//   p = PrivacyPublic,  pb = PrivacyPublic  -> true
//   p = PrivacyPrivate, pb = PrivacyPublic  -> true
func (p Privacy) CanAccess(pb Privacy) bool {
	return p >= pb
}

func (p Privacy) String() string {
	switch p {
	case PrivacyPublic:
		return "public"
	case PrivacyProtected:
		return "protected"
	case PrivacyPrivate:
		return "private"
	default:
		panic("Unknown privacy given")
	}
}

func FromString(p string) (Privacy, error) {
	switch p {
	case "public":
		return PrivacyPublic, nil
	case "protected":
		return PrivacyProtected, nil
	case "private":
		return PrivacyPrivate, nil
	default:
		return -1, fmt.Errorf("%s is not a privacy value", p)
	}
}
