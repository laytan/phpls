package phpversion_test

import (
	"testing"

	"github.com/laytan/elephp/pkg/phpversion"
)

func TestPHPVersion_String(t *testing.T) {
	t.Parallel()
	type fields struct {
		Major uint8
		Minor uint8
		Patch uint8
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "nothing",
			fields: fields{},
			want:   "0.0.0",
		},
		{
			name:   "normal",
			fields: fields{Major: 7, Minor: 3, Patch: 1},
			want:   "7.3.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := &phpversion.PHPVersion{
				Major: tt.fields.Major,
				Minor: tt.fields.Minor,
				Patch: tt.fields.Patch,
			}
			if got := v.String(); got != tt.want {
				t.Errorf("PHPVersion.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
