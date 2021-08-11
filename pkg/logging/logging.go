package logging

var GlobaLogger *Log = New(LevelInfo, FormatConsole)

type Config struct {
	Level  string `envconfig:"LOG_LEVEL" default:"info"`
	Format string `envconfig:"LOG_FORMAT" default:"json"`
}

func (c Config) Apply(log *Log) {
	SetLevel(GetLevelByName(c.Level)).Apply(log)
	SetFormatter(GetFormatterByName(c.Format)).Apply(log)
}

type Meta map[string]interface{}

type Logger interface {
	Debug(message string, fields ...interface{})
	Info(message string, fields ...interface{})
	Warn(message string, fields ...interface{})
	Error(message string, err error, fields ...interface{})
}

type Option interface {
	Apply(*Log)
}

type OptionFunc func(*Log)

func (f OptionFunc) Apply(log *Log) {
	f(log)
}

func SetLevel(level Level) Option {
	return OptionFunc(func(log *Log) {
		log.level = level
	})
}

func SetFormatter(formatter Formatter) Option {
	return OptionFunc(func(log *Log) {
		log.formatter = formatter
	})
}
