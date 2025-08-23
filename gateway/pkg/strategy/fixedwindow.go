package strategy

import (
	"context"
	"database/sql"
	"fmt"
	errorlog "gateway/pkg/error-log"
	"time"
)

// FixedWindow strategy uses an SQL table to track request counts per user and path.
type FixedWindow struct {
	LengthSeconds int
	SqlDb         *sql.DB
	Logger        *errorlog.Logger
	SqlTable      string
}

// Accept checks if a request from a user for a specific path is allowed based on the fixed window strategy.
// The algorithm:
// - checks if there is an open window
// - if there is, the window is used to check if the request can be accepted
// - if there isn't, a new window is created. Older windows are deleted.
func (fw *FixedWindow) Accept(userId string, requestsPerSecond float64, path string) bool {

	nowSeconds := time.Now().Unix()
	currentWindowStart := nowSeconds - (nowSeconds % int64(fw.LengthSeconds))
	maxRequests := int(requestsPerSecond * float64(fw.LengthSeconds))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start a transaction
	tx, err := fw.SqlDb.BeginTx(ctx, nil)
	if err != nil {
		fw.Logger.WriteError(fmt.Errorf("failed to begin tx: %w", err))
		return false
	}
	defer tx.Rollback()

	var count int
	err = tx.QueryRowContext(ctx, `
		SELECT count FROM `+fw.SqlTable+`
		WHERE user_id = ? AND path = ? AND window_start = ?`,
		userId, path, currentWindowStart,
	).Scan(&count)

	switch err {
	case nil:
		if count >= maxRequests {
			return false
		}
		_, err = tx.ExecContext(ctx, `
			UPDATE `+fw.SqlTable+`
			SET count = count + 1
			WHERE user_id = ? AND path = ? AND window_start = ?`,
			userId, path, currentWindowStart,
		)
		if err != nil {
			fw.Logger.WriteError(fmt.Errorf("sql update failed: %w", err))
			return false
		}
	case sql.ErrNoRows:
		fw.SqlDb.Exec("DELETE FROM "+fw.SqlTable+" WHERE window_start < ?", currentWindowStart)

		_, err = tx.ExecContext(ctx, `
			INSERT INTO `+fw.SqlTable+` (user_id, path, window_start, count)
			VALUES (?, ?, ?, 1)`,
			userId, path, currentWindowStart,
		)
		if err != nil {
			fw.Logger.WriteError(fmt.Errorf("sql insert failed: %w", err))
			return false
		}

	default:
		fw.Logger.WriteError(fmt.Errorf("sql query failed: %w", err))
		return false
	}

	if err := tx.Commit(); err != nil {
		fw.Logger.WriteError(fmt.Errorf("sql commit failed: %w", err))
		return false
	}

	return true

}
