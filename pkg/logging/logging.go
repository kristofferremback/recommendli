package logging

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
	Debug(message string)
	Info(message string)
	Warn(message string)
	Error(message string, err error)
	Fatal(message string, err error)
	WithFields(fields Meta) Logger
	With(key string, value interface{}) Logger
	WithOptions(opts ...Option) Logger
	Sync() error
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
