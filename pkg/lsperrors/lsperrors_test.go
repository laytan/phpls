package lsperrors_test

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/matryer/is"
)

func TestErrorCodes(t *testing.T) {
	t.Parallel()
	is := is.New(t)

	expectations := map[int]error{
		-32002: lsperrors.ErrServerNotInitialized,
		-32803: lsperrors.ErrRequestFailed(""),
		-32802: lsperrors.ErrServerCancelled,
		-32801: lsperrors.ErrContentModified,
		-32800: lsperrors.ErrRequestCancelled,
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
