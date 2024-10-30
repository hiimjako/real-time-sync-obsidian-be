package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAuthenticated(t *testing.T) {
	ao := AuthOptions{SecretKey: []byte("secret-key")}

	createToken := func(workspaceID int) string {
		token, err := CreateToken(ao, workspaceID)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		return token
	}

	tests := []struct {
		name                string
		authHeader          string
		expectedStatus      int
		expectedWorkspaceID int
	}{
		{"No Auth Header", "", http.StatusUnauthorized, 0},
		{"Invalid Token", "Bearer invalidToken", http.StatusUnauthorized, 0},
		{"Valid Token", "Bearer " + createToken(123), http.StatusOK, 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rec := httptest.NewRecorder()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tt.expectedWorkspaceID, WorkspaceIDFromCtx(r.Context()))
				w.WriteHeader(http.StatusOK)
			})

			handler := IsAuthenticated(ao)(next)
			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestWorkspaceIDFromCtx(t *testing.T) {
	expectedWorkspaceID := 10
	ctx := context.WithValue(context.Background(), AuthWorkspaceID, 10)
	workspaceID := WorkspaceIDFromCtx(ctx)

	assert.Equal(t, expectedWorkspaceID, workspaceID)
}
