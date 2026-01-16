package cmdx

import (
	"context"
	"fmt"
	"os"

	"github.com/cnk3x/xunlei/pkg/utils"
)

func ShellFile(ctx context.Context, file string, options ...CmdOption) error {
	if !utils.PathExists(file) {
		return fmt.Errorf("%w: %s", os.ErrNotExist, file)
	}
	return Exec(ctx, "sh", append([]CmdOption{Args(file), Std()}, options...)...)
}
