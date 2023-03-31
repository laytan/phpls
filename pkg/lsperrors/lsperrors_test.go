package lsperrors_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/laytan/elephp/pkg/lsperrors"
	"github.com/stretchr/testify/require"
)

func TestErrorCodes(t *testing.T) {
	t.Parallel()

	expectations := map[int]error{
		-32002: lsperrors.ErrServerNotInitialized,
		-32803: lsperrors.ErrRequestFailed(""),
		-32802: lsperrors.ErrServerCancelled,
		-32801: lsperrors.ErrContentModified,
		-32800: lsperrors.ErrRequestCancelled,
	}

	for code, err := range expectations {
		json, jsonErr := json.Marshal(err)
		require.NoError(t, jsonErr)
		require.Contains(t, string(json), strconv.Itoa(code))
	}
}
