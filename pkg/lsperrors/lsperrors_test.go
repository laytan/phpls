package lsperrors

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/matryer/is"
)

func TestErrorCodes(t *testing.T) {
	is := is.New(t)

	expectations := map[int]error{
		-32002: ErrServerNotInitialized,
		-32803: ErrRequestFailed(""),
		-32802: ErrServerCancelled,
		-32801: ErrContentModified,
		-32800: ErrRequestCancelled,
	}

	for code, err := range expectations {
		json, jsonErr := json.Marshal(err)
		is.NoErr(jsonErr)

		contains := strings.Contains(string(json), strconv.Itoa(code))
		if !contains {
			t.Errorf("Expected the lsp error %s to contain the code %d", string(json), code)
		}
	}
}
