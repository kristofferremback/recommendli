package logging

import (
	"fmt"
	"strings"
)

type Level int

const (
	LevelDebug Level = iota - 1
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal

	_defaultLevel = LevelInfo
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	default:
		return fmt.Sprintf("Level(%d)", l)
	}
}

func (l Level) Apply(log *Log) {
	log.level = l
}

func GetLevelByName(name string) Level {
	if level, ok := map[string]Level{
		"debug": LevelDebug,
		"info":  LevelInfo,
		"warn":  LevelWarn,
		"error": LevelError,
		"fatal": LevelFatal,
	}[strings.ToLower(name)]; ok {
		return level
	}

	return _defaultLevel
}
