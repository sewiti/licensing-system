package core

import (
	"context"
	"fmt"
	"time"

	"github.com/sewiti/licensing-system/internal/db"
)

type CleanupCallback func(msg string, err error)

func (cb CleanupCallback) call(msg string, err error) {
	if cb != nil {
		cb(msg, err)
	}
}

// RunCleanupRoutine runs license sessions cleaner routine. This routine
// periodically cleans up expired and overused license sessions from the
// database.
//
// Calls callback with cleanup info and an error if any
// (nil error means deletion report).
//
// Blocks until context is canceled.
func RunCleanupRoutine(ctx context.Context, db *db.Handler, interval time.Duration, cb CleanupCallback) {
	cleanup(ctx, db, cb)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			cleanup(ctx, db, cb)
		case <-ctx.Done():
			return
		}
	}
}

// cleanup deletes expired and overused license sessions.
//
// Calls callback with info about deletion and an error if any.
func cleanup(ctx context.Context, db *db.Handler, cb CleanupCallback) {
	n, err := db.DeleteLicenseSessionsExpiredBy(ctx, time.Now())
	if err != nil {
		cb.call("deleting expired license sessions", err)
	} else {
		cb.call(fmt.Sprintf("deleted %d expired license sessions", n), nil)
	}

	n, err = db.DeleteLicenseSessionsOverused(ctx)
	if err != nil {
		cb.call("deleting overused license sessions", err)
	} else {
		cb.call(fmt.Sprintf("deleted %d overused license sessions", n), nil)
	}
}
