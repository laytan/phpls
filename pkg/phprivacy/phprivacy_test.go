package phprivacy

import "testing"

func TestPrivacy_CanAccess(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		p    Privacy
		pb   Privacy
		want bool
	}{
		{
			name: "Public -> Private",
			p:    PrivacyPublic,
			pb:   PrivacyPrivate,
			want: false,
		},
		{
			name: "Private -> Public",
			p:    PrivacyPrivate,
			pb:   PrivacyPublic,
			want: true,
		},
		{
			name: "Protected -> Public",
			p:    PrivacyProtected,
			pb:   PrivacyPublic,
			want: true,
		},
		{
			name: "Private -> Private",
			p:    PrivacyPrivate,
			pb:   PrivacyPrivate,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.p.CanAccess(tt.pb); got != tt.want {
				t.Errorf("Privacy.CanAccess() = %v, want %v", got, tt.want)
			}
		})
	}
}
