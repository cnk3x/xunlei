package lod

import "log/slog"

// ErrDebug 返回日志级别，如果err不为零，返回 slog.LevelDebug，否则返回 slog.LevelInfo
func ErrDebug(err error) slog.Level {
	return Iif(err == nil, slog.LevelDebug, slog.LevelWarn)
}

// ErrInfo 返回日志级别，如果err不为零，返回 slog.LevelInfo，否则返回 slog.LevelInfo
func ErrInfo(err error) slog.Level {
	return Iif(err == nil, slog.LevelInfo, slog.LevelWarn)
}
