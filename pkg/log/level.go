package log

import (
	"cmp"
	"log/slog"
	"strings"
)

// Level sets the log level
func Level(logger *slog.Logger, level string) {
	if h, ok := logger.Handler().(*handler); ok {
		var def slog.Level = slog.LevelInfo
		if h.level != nil {
			def = h.level.Level()
		}
		h.level = LevelFromString(level, def)
	}
}

// LevelFromString parse the level from string, ignore case
//   - debug => slog.LevelDebug
//   - info, information => slog.LevelInfo
//   - warn, warning => slog.LevelWarn
//   - error, err => slog.LevelError
//   - otherwise slog.LevelInfo
func LevelFromString(level string, def slog.Level) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info", "information":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "err":
		return slog.LevelError
	default:
		return def
	}
}

// ForDefault 函数用于设置默认日志记录器的级别和是否添加源信息。
//
// 参数:
//   - level string - 日志级别的字符串表示，例如 "info", "debug" 等。
//   - addSource ...bool - 可选参数，用于指定是否添加源信息。如果提供，第一个布尔值将被用于设置是否添加源信息。否则根据 level == debug 判断添加源信息。
func ForDefault(level string, addSource ...bool) {
	l := LevelFromString(level, slog.LevelInfo)
	// 初始化选项，当日志级别为Debug时
	lOpt := &Options{Level: l, AddSource: cmp.Or(addSource...)}
	// 设置默认日志记录器
	slog.SetDefault(New(lOpt))
}
