package phprivacy_test

import (
	"testing"

	"github.com/laytan/phpls/pkg/phprivacy"
)

func TestPrivacy_CanAccess(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		p    phprivacy.Privacy
		pb   phprivacy.Privacy
		want bool
	}{
		{
			name: "Public -> Private",
			p:    phprivacy.PrivacyPublic,
			pb:   phprivacy.PrivacyPrivate,
			want: false,
		},
		{
			name: "Private -> Public",
			p:    phprivacy.PrivacyPrivate,
			pb:   phprivacy.PrivacyPublic,
			want: true,
		},
		{
			name: "Protected -> Public",
			p:    phprivacy.PrivacyProtected,
			pb:   phprivacy.PrivacyPublic,
			want: true,
		},
		{
			name: "Private -> Private",
			p:    phprivacy.PrivacyPrivate,
			pb:   phprivacy.PrivacyPrivate,
			want: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.p.CanAccess(tt.pb); got != tt.want {
				t.Errorf("phprivacy.Privacy.CanAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}
