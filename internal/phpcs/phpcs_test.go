package phpcs_test

import (
	"testing"

	"github.com/laytan/elephp/internal/phpcs"
	"github.com/laytan/go-lsp-protocol/pkg/lsp/protocol"
	"github.com/stretchr/testify/require"
)

func TestFormatFile(t *testing.T) {
	t.Parallel()
	type tcase struct {
		name     string
		code     string
		expected []protocol.TextEdit
	}
	cases := []tcase{
		{
			name: "simple start",
			code: `<?php
class Test {}`,
			expected: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0},
					End:   protocol.Position{Line: 2},
				},
				NewText: `<?php

class Test
{
}
`,
			}},
		},
		{
			name: "array edit",
			code: `<?php
add('shared_files', [
  '.env',
  'data/database.sqlite',
]);
`,
			expected: []protocol.TextEdit{{
				Range: protocol.Range{
					Start: protocol.Position{Line: 0},
					End:   protocol.Position{Line: 4},
				},
				NewText: `<?php

add('shared_files', [
  '.env',
  'data/database.sqlite',
`,
			}},
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			eds, err := phpcs.FormatCodeEdits([]byte(tt.code))
			require.NoError(t, err)
			require.Equal(t, tt.expected, eds)
		})
	}
}
