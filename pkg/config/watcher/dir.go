package watcher

import (
	"context"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type Dir struct {
	Path string
}

func (f Dir) Watch(ctx context.Context, reload func() error) error {
	// Resolve symlinks and save the original path so that changes to symlinks
	// can be detected.
	realPath, err := filepath.EvalSymlinks(f.Path)
	if err != nil {
		return err
	}
	realPath = filepath.Clean(realPath)
	fDir, _ := filepath.Split(f.Path)

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	var (
		lastEvent     string
		lastEventTime time.Time
	)

	err = w.Add(fDir)
	if err != nil {
		return errors.Wrap(err, "unable to add watch dir")
	}

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return errors.New("fsnotify watch channel closed")
			}

			// Use a simple timer to buffer events as certain events fire
			// multiple times on some platforms.
			if event.String() == lastEvent && time.Since(lastEventTime) < time.Millisecond*5 {
				continue
			}
			lastEvent = event.String()
			lastEventTime = time.Now()

			// Resolve symlink to get the real path, in case the symlink's
			// target has changed.
			curPath, err := filepath.EvalSymlinks(f.Path)
			if err != nil {
				return err
			}
			realPath = filepath.Clean(curPath)

			// Finally, we only care about create and write.
			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Remove) == 0 {
				continue
			}

			// Trigger event.
			if err = reload(); err != nil {
				return err
			}

		// There's an error.
		case err, ok := <-w.Errors:
			if !ok {
				return errors.New("fsnotify err channel closed")
			}

			return err
		case <-ctx.Done():
			return nil
		}
	}
}
