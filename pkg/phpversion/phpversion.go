package phpversion

import (
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	versionStringPartsLength = 3
	versionBase              = 10
	versionBitSize           = 8
)

var versionRgx = regexp.MustCompile(`^(\d+)\.?(\d*)\.?(\d*)$`)

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

func Get() (*PHPVersion, error) {
	cmd := exec.Command("php", "-v")
	output, err := cmd.Output()
	if err != nil && reflect.TypeOf(err) != reflect.TypeOf(exec.ExitError{}) {
		return nil, fmt.Errorf("Error parsing 'php -v' output: %w", err)
	}

	parts := strings.Split(string(output), " ")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Unexpected output from 'php -v': %s", string(output))
	}

	versionString := strings.Split(parts[1], ".")
	if len(versionString) != versionStringPartsLength {
		return nil, fmt.Errorf("Unexpected output version from 'php -v': %s", parts[1])
	}

	major, err := strconv.ParseUint(versionString[0], versionBase, versionBitSize)
	if err != nil {
		return nil, versionConversionErr(err)
	}

	minor, err := strconv.ParseUint(versionString[1], versionBase, versionBitSize)
	if err != nil {
		return nil, versionConversionErr(err)
	}

	patch, err := strconv.ParseUint(versionString[2], versionBase, versionBitSize)
	if err != nil {
		return nil, versionConversionErr(err)
	}

	return &PHPVersion{
		Major: uint8(major),
		Minor: uint8(minor),
		Patch: uint8(patch),
	}, nil
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

func versionConversionErr(err error) error {
	return fmt.Errorf("Unexpected php version conversion error: %w", err)
}
