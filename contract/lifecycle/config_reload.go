package lifecycle

import (
	"context"

	"github.com/DoNewsCode/core/contract"
)

type ConfigReload interface {
	Fire(ctx context.Context, Config contract.ConfigUnmarshaler) error
	On(func(ctx context.Context, Config contract.ConfigUnmarshaler) error) func()
}
