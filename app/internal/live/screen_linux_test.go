package live

import "testing"

func Test_getPipewireNodeID(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		applicationPID uint
		want           uint
		wantErr        bool
	}{
		{
			name:           "local",
			applicationPID: 21966,
			want:           62,
			wantErr:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := getPipewireNodeID(tt.applicationPID)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getPipewireNodeID() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getPipewireNodeID() succeeded unexpectedly")
			}
			if got != tt.want {
				t.Errorf("getPipewireNodeID() = %v, want %v", got, tt.want)
			}
		})
	}
}
