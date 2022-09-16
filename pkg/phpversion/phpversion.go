package phpversion

import (
	"fmt"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

const (
	versionStringPartsLength = 3
	versionBase              = 10
	versionBitSize           = 8
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

func versionConversionErr(err error) error {
	return fmt.Errorf("Unexpected php version conversion error: %w", err)
}
