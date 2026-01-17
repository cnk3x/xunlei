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
	return Exec(ctx, "sh", append([]Option{Args(file), LogStd()}, options...)...)
}

func Shell(ctx context.Context, shellScript []string, options ...Option) error {
	return Exec(ctx, "sh", append(options, SlogDebug("cmdx"), LineIn(shellScript...))...)
}
