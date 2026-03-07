package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRepository(t *testing.T) {
	// postgresURL, ok := os.LookupEnv("POSTGRES_URL")
	// if !ok {
	// 	t.Fatal("mandatory environment variable POSTGRES_URL not set")
	// }

	// invalidURL := "postgres://invalid:invalid@localhost:0/nonexistent?sslmode=disable"

	// type args struct {
	// 	dbURL string
	// }

	tests := []struct {
		name string
		// args    args
		wantErr bool
	}{
		{
			name: "happy case: successfully connect to database",
			// args:    args{dbURL: postgresURL},
			wantErr: false,
		},
		{
			name: "sad case: invalid connection string",
			// args:    args{dbURL: invalidURL},
			wantErr: true,
		},
		{
			name: "sad case: empty connection string",
			// args:    args{dbURL: ""},
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
