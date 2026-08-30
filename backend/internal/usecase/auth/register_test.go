package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRegisterInput(t *testing.T) {
	tests := []struct {
		name    string
		input   RegisterInput
		wantErr bool
	}{
		{
			name: "valid input",
			input: RegisterInput{
				Username:    "alice_dev",
				Email:       "alice@example.com",
				Password:    "password123",
				DisplayName: "Alice Developer",
			},
			wantErr: false,
		},
		{
			name: "username too short",
			input: RegisterInput{
				Username:    "al",
				Email:       "alice@example.com",
				Password:    "password123",
				DisplayName: "Alice",
			},
			wantErr: true,
		},
		{
			name: "username invalid characters",
			input: RegisterInput{
				Username:    "alice-dev!",
				Email:       "alice@example.com",
				Password:    "password123",
				DisplayName: "Alice",
			},
			wantErr: true,
		},
		{
			name: "password too short",
			input: RegisterInput{
				Username:    "alice_dev",
				Email:       "alice@example.com",
				Password:    "short",
				DisplayName: "Alice",
			},
			wantErr: true,
		},
		{
			name: "email missing @",
			input: RegisterInput{
				Username:    "alice_dev",
				Email:       "aliceexample.com",
				Password:    "password123",
				DisplayName: "Alice",
			},
			wantErr: true,
		},
		{
			name: "display name too short",
			input: RegisterInput{
				Username:    "alice_dev",
				Email:       "alice@example.com",
				Password:    "password123",
				DisplayName: "A",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRegisterInput(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}