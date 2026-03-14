package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "happy case: successfully connect to database",
			wantErr: false,
		},
		{
			name:    "sad case: invalid connection string",
			wantErr: true,
		},
		{
			name:    "sad case: empty connection string",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New()

			if (err != nil) != tt.wantErr {
				t.Errorf("TestNewRepository got %v, want %v", err, tt.wantErr)
			}

			if tt.name == "happy case: successfully connect to database" {
				assert.NotNilf(t, r, "expected non-nil repository")
			}

			if tt.name == "sad case: invalid connection string" {
				assert.Nilf(t, r, "expected nil repository")
			}
		})
	}
}
