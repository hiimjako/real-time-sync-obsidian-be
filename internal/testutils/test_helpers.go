package testutils

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/migration"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CreateDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NoError(t, migration.Migrate(db))

	t.Cleanup(func() { db.Close() })

	return db
}

type requestOption func(req *http.Request) error

func WithAuthHeader(secretKey []byte, workspaceID int64) requestOption {
	return func(req *http.Request) error {
		token, err := middleware.CreateToken(middleware.AuthOptions{SecretKey: secretKey}, workspaceID)
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", "Bearer "+token)
		return nil
	}
}

func DoRequest[T any](
	t *testing.T,
	server http.Handler,
	path string,
	input any,
	options ...requestOption,
) (*httptest.ResponseRecorder, T) {
	reqBody, err := json.Marshal(input)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(reqBody))
	for _, opt := range options {
		require.NoError(t, opt(req))
	}

	res := httptest.NewRecorder()

	server.ServeHTTP(res, req)

	var resBody T
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	// If T is string, return the body as a string
	if str, ok := any(&resBody).(*string); ok {
		if body == nil {
			*str = ""
		} else {
			*str = strings.Trim(string(body), "\n")
		}
	} else {
		err = json.Unmarshal(body, &resBody)
		assert.NoError(t, err)
	}

	return res, resBody
}
