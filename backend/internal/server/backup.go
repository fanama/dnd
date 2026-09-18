package server

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"dnd-backend/internal/repository"
)

// SnapshotDB copies the SQLite database (and its WAL files, if any) into
// dir/game-<timestamp>.db after a WAL checkpoint, then prunes old backups so
// at most keep snapshots survive.
func SnapshotDB(repo *repository.SQLiteRepository, dbPath, dir string, keep int, logger *slog.Logger) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	repo.Checkpoint()

	name := "game-" + time.Now().Format("20060102-150405") + ".db"
	dst := filepath.Join(dir, name)

	sources := []string{dbPath, dbPath + "-wal", dbPath + "-shm"}
	for _, src := range sources {
		in, err := os.Open(src)
		if err != nil {
			continue
		}
		out, err := os.Create(dst + strings.TrimPrefix(src, dbPath))
		if err != nil {
			in.Close()
			return "", err
		}
		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			out.Close()
			return "", err
		}
		in.Close()
		out.Close()
	}

	pruneBackups(dir, keep, logger)
	return dst, nil
}

// pruneBackups removes the oldest game-*.db snapshots beyond the keep limit.
func pruneBackups(dir string, keep int, logger *slog.Logger) {
	matches, err := filepath.Glob(filepath.Join(dir, "game-*.db*"))
	if err != nil {
		return
	}
	sort.Strings(matches)
	for len(matches) > keep {
		old := matches[0]
		if err := os.Remove(old); err != nil {
			logger.Warn("backup prune failed", "file", old, "err", err)
		} else {
			logger.Info("backup pruned", "file", old)
		}
		matches = matches[1:]
	}
}

// StartBackupRoutine backs up on startup and then every interval until ctx is
// cancelled. A final backup runs on shutdown to capture late writes.
func StartBackupRoutine(ctx context.Context, repo *repository.SQLiteRepository, dbPath, dir string, keep int, interval time.Duration, logger *slog.Logger) {
	if interval <= 0 {
		interval = 15 * time.Minute
	}
	// Startup backup.
	if p, err := SnapshotDB(repo, dbPath, dir, keep, logger); err != nil {
		logger.Warn("startup backup failed", "err", err)
	} else {
		logger.Info("backup created", "file", p)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			if p, err := SnapshotDB(repo, dbPath, dir, keep, logger); err != nil {
				logger.Warn("shutdown backup failed", "err", err)
			} else {
				logger.Info("shutdown backup created", "file", p)
			}
			return
		case <-ticker.C:
			if p, err := SnapshotDB(repo, dbPath, dir, keep, logger); err != nil {
				logger.Warn("scheduled backup failed", "err", err)
			} else {
				logger.Info("backup created", "file", p)
			}
		}
	}
}