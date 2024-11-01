package testutils

import (
	"database/sql"
	"net/http"
	"testing"

	"github.com/hiimjako/real-time-sync-obsidian-be/internal/migration"
	"github.com/hiimjako/real-time-sync-obsidian-be/pkg/middleware"
	"github.com/stretchr/testify/require"
)

func CreateAuthHeader(req *http.Request, secretKey []byte, userID int64) error {
	token, err := middleware.CreateToken(middleware.AuthOptions{SecretKey: secretKey}, userID)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	return nil
}

func CreateDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	require.NoError(t, migration.Migrate(db))

	t.Cleanup(func() { db.Close() })

	return db
}
