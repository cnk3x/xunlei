package cmdx

import (
	"context"
	"fmt"
	"os"

	"github.com/cnk3x/xunlei/pkg/utils"
)

func FileShell(ctx context.Context, file string, options ...Option) error {
	if !utils.PathExists(file) {
		return fmt.Errorf("%w: %s", os.ErrNotExist, file)
	}
	return Run(ctx, "sh", append([]Option{Args(file), LogStd()}, options...)...)
}

func MultiShell(ctx context.Context, shellScript []string, options ...Option) error {
	return Run(ctx, "sh", append(options, LineIn(shellScript...))...)
}

func Shell(ctx context.Context, shellScript string, options ...Option) error {
	return Run(ctx, "sh", append(options, Args("-c", shellScript))...)
}
