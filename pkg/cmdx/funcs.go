package cmdx

import (
	"context"
)

func MultiShell(ctx context.Context, shellScript []string, options ...Option) error {
	return Run(ctx, "sh", append(options, LineIn(shellScript...))...)
}

func Shell(ctx context.Context, shellScript string, options ...Option) error {
	return Run(ctx, "sh", append(options, Args("-c", shellScript))...)
}
