package phpversion

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

var (
	versionRgx    = regexp.MustCompile(`^(\d+)\.?(\d*)\.?(\d*)$`)
	cliVersionRgx = regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)
)

// PHPVersion struct is a representation of a PHP version with methods to retrieve it.
type PHPVersion struct {
	Major uint8
	Minor uint8
	Patch uint8
}

func (v *PHPVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func (v *PHPVersion) IsHigherThan(other *PHPVersion) bool {
	if v.Major > other.Major {
		return true
	}

	if other.Major > v.Major {
		return false
	}

	if v.Minor > other.Minor {
		return true
	}

	if other.Minor > v.Minor {
		return false
	}

	if v.Patch > other.Patch {
		return true
	}

	if other.Patch > v.Patch {
		return false
	}

	// Same version.
	return false
}

func (v *PHPVersion) Equals(other *PHPVersion) bool {
	return v.Major == other.Major && v.Minor == other.Minor && v.Patch == other.Patch
}

func (v *PHPVersion) EqualsMajorMinor(other *PHPVersion) bool {
	return v.Major == other.Major && v.Minor == other.Minor
}

func EightOne() *PHPVersion {
	return &PHPVersion{Major: 8, Minor: 1}
}

func FromString(version string) (*PHPVersion, bool) {
	match := versionRgx.FindStringSubmatch(version)
	if len(match) < 2 {
		return nil, false
	}

	// NOTE: Errors won't happen because the regex makes sure these are valid integers.

	major, _ := strconv.ParseUint(match[1], 10, 8)

	var minor uint8
	if match[2] != "" {
		m, _ := strconv.ParseUint(match[2], 10, 8)
		minor = uint8(m)
	}

	var patch uint8
	if match[3] != "" {
		p, _ := strconv.ParseUint(match[3], 10, 8)
		patch = uint8(p)
	}

	return &PHPVersion{
		Major: uint8(major),
		Minor: minor,
		Patch: patch,
	}, true
}

func Get() (*PHPVersion, error) {
	cmd := exec.Command("php", "-v")
	out, err := cmd.Output()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		return nil, fmt.Errorf("running php -v: %w", err)
	}

	matches := cliVersionRgx.FindSubmatch(out)
	if len(matches) != 2 {
		return nil, fmt.Errorf("could not parse php version from output \"%s\"", string(out))
	}

	version, ok := FromString(string(matches[1]))
	if !ok {
		return nil, fmt.Errorf("could not parse %s into a version", string(matches[1]))
	}

	return version, nil
}
