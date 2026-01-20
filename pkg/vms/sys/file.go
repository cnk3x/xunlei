package sys

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cnk3x/xunlei/pkg/fo"
	"github.com/cnk3x/xunlei/pkg/utils"
)

func NewFile(ctx context.Context, fn string, process fo.Process, fopts ...fo.Option) (undo func(), err error) {
	bq := utils.BackQueue(&undo, &err)
	defer bq.ErrDefer()

	mdUndo, e := Mkdir(ctx, filepath.Dir(fn), 0777)
	if err = e; err != nil {
		err = fmt.Errorf("mkdir for create file: %w", e)
		return
	}
	bq.Put(mdUndo)

	if err = fo.OpenWrite(fn, process, append(fopts, fo.FlagExcl(false))...); err == nil {
		bq.Put(newRm(ctx, fn, "rmfile"))
	}

	if exists := os.IsExist(err); err != nil && exists {
		err = nil
	}

	return
}
