package otredis

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/DoNewsCode/core/logging"
	"github.com/go-kit/log"
	"github.com/spf13/cobra"
)

// NewCleanupCommand creates a new command to clean up unused redis keys.
func NewCleanupCommand(maker Maker, baseLogger log.Logger) *cobra.Command {
	const cursorFileName = "redis-scan-cursor"
	var (
		logger     = logging.WithLevel(baseLogger)
		cursorPath string
		batchSize  int64
		prefix     string
		instance   string
	)

	type stats struct {
		scanned uint64
		removed uint64
	}

	removeKeys := func(ctx context.Context, cursor *uint64, stats *stats, threshold time.Duration) error {
		var (
			keys []string
			err  error
		)

		logger.Info(fmt.Sprintf("scanning redis keys from cursor %d", *cursor))
		if err := os.WriteFile(cursorPath, []byte(fmt.Sprintf("%d", *cursor)), os.ModePerm); err != nil {
			logger.Err("cannot store cursor to cursor location", err, nil)
		}

		redisClient, err := maker.Make(instance)
		if err != nil {
			return fmt.Errorf("cannot find redis instance under the name of %s: %w", instance, err)
		}
		keys, *cursor, err = redisClient.Scan(ctx, *cursor, prefix+"*", batchSize).Result()
		if err != nil {
			return err
		}
		stats.scanned += uint64(len(keys))

		var wg sync.WaitGroup
		for _, key := range keys {
			wg.Add(1)
			go func(key string) {
				idleTime, _ := redisClient.ObjectIdleTime(ctx, key).Result()
				if idleTime > threshold {
					logger.Info(fmt.Sprintf("removing %s from redis as it is %s old", key, idleTime))
					redisClient.Del(ctx, key)
					atomic.AddUint64(&stats.removed, 1)
				}
				wg.Done()
			}(key)
		}
		wg.Wait()
		return nil
	}

	initCursor := func() (uint64, error) {
		cursorPath = filepath.Join(os.TempDir(), fmt.Sprintf("%s.%s.txt", cursorFileName, instance))
		if _, err := os.Stat(cursorPath); os.IsNotExist(err) {
			return 0, nil
		}
		f, err := os.Open(cursorPath)
		if err != nil {
			return 0, err
		}
		defer f.Close()
		lastCursor, err := io.ReadAll(f)
		if err != nil {
			return 0, err
		}
		cursor, err := strconv.ParseUint(string(lastCursor), 10, 64)
		if err != nil {
			return 0, err
		}
		return cursor, nil
	}

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "clean up idle keys in redis",
		Args:  cobra.ExactArgs(1),
		Long:  `clean up idle keys in redis based on the last access time.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				cursor   uint64
				stats    stats
				duration time.Duration
			)

			defer func() {
				if stats.scanned == 0 {
					return
				}
				logger.Info(fmt.Sprintf("%d keys scanned, %d keys removed(%.2f%%)", stats.scanned, stats.removed, float64(stats.removed)/float64(stats.scanned)), nil)
			}()

			duration, err := time.ParseDuration(args[0])
			if err != nil {
				return fmt.Errorf("first argument must be valid duration string, got %w", err)
			}

			cursor, err = initCursor()
			if err != nil {
				return fmt.Errorf("error restoring cursor from file: %w", err)
			}

			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

			for {
				select {
				case <-cmd.Context().Done():
					return cmd.Context().Err()
				case <-shutdown:
					return nil
				default:
					if err := removeKeys(cmd.Context(), &cursor, &stats, duration); err != nil {
						return fmt.Errorf("error while removing keys: %w", err)
					}
				}
				if cursor == 0 {
					break
				}
			}
			logger.Info("redis key clean-up completed", nil)
			os.Remove(cursorPath)
			return nil
		},
	}

	cmd.Flags().Int64VarP(&batchSize, "batchSize", "b", 100, "specify the redis scan batch size")
	cmd.Flags().StringVarP(&cursorPath, "cursorPath", "c", cursorPath, "specify the location to store the cursor, so that the next execution can continue from where it's left off.")
	cmd.Flags().StringVarP(&prefix, "prefix", "p", "", "specify the prefix of redis keys to be scanned")
	cmd.Flags().StringVarP(&instance, "instance", "i", "default", "specify the redis instance to be scanned")

	return cmd
}
