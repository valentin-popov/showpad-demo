package limiter

import (
	"context"
	"database/sql"

	_ "modernc.org/sqlite"
)

// NewDB initializes and returns a new database connection.
func NewDB(ctx context.Context, filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:"+filename)
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}

func (l *Limiter) updateUserQuota(userId string, requestRate float64) error {
	_, err := l.sqlDb.Exec(`
	UPDATE users SET quota = ? WHERE id = ?`,
		requestRate, userId,
	)

	if err == nil {
		l.userIdCache.Add(userId, requestRate)
	}

	return err
}
