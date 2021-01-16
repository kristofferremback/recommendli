package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/errors"
)

type WriteSyncer interface {
	io.Writer
	Sync() error
}

type Log struct {
	level     Level
	formatter Formatter
	meta      Meta

	transport WriteSyncer
}

type Message struct {
	Level     Level
	Body      string
	Timestamp time.Time
	Meta      Meta
	Caller    Caller // @TODO: Implement and/or use some form of caller
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func New(opts ...Option) Logger {
	log := &Log{
		level:     _defaultLevel,
		formatter: _defaultFormatter,
		meta:      Meta{},

		transport: os.Stdout,
	}

	for _, opt := range opts {
		opt.Apply(log)
	}

	return log
}

func (l *Log) log(level Level, message string, meta Meta) {
	if level < l.level {
		return
	}

	if meta == nil {
		meta = Meta{}
	}

	msg := &Message{
		Level:     level,
		Body:      message,
		Meta:      meta,
		Timestamp: time.Now(),
		Caller:    getCaller(4),
	}

	for key, value := range l.meta {
		if _, hasOverride := msg.Meta[key]; !hasOverride {
			msg.Meta[key] = value
		}
	}

	out := l.formatter(msg)
	fmt.Fprint(l.transport, string(out))
}

func (l *Log) Debug(message string) {
	l.log(LevelDebug, message, nil)
}

func (l *Log) Info(message string) {
	l.log(LevelInfo, message, nil)
}

func (l *Log) Warn(message string) {
	l.log(LevelWarn, message, nil)
}

func (l *Log) Error(message string, err error) {
	l.log(LevelError, message, l.metaWithError(err))
}

func (l *Log) Fatal(message string, err error) {
	defer os.Exit(1)
	defer l.Sync()

	l.log(LevelFatal, message, l.metaWithError(err))
}

func (l *Log) WithFields(fields Meta) Logger {
	log := &Log{
		level:     l.level,
		formatter: l.formatter,
		transport: l.transport,
		meta:      Meta{},
	}

	// Clone existing meta map
	for key, value := range l.meta {
		log.meta[key] = value
	}

	for key, value := range fields {
		log.meta[key] = value
	}

	return log
}

func (l *Log) With(key string, value interface{}) Logger {
	log := &Log{
		level:     l.level,
		formatter: l.formatter,
		transport: l.transport,
		meta:      Meta{},
	}

	// Clone existing meta map
	for k, value := range l.meta {
		log.meta[k] = value
	}
	log.meta[key] = value

	return log
}

func (l *Log) WithOptions(opts ...Option) Logger {
	log := &Log{
		level:     l.level,
		formatter: l.formatter,
		transport: l.transport,
		meta:      Meta{},
	}

	// Clone existing meta map
	for key, value := range l.meta {
		log.meta[key] = value
	}

	for _, opt := range opts {
		opt.Apply(log)
	}

	return log
}

func (l *Log) Sync() error {
	err := l.transport.Sync()
	if err != nil {
		return errors.Wrap(err, "Failed to sync transport")
	}

	return nil
}

func (l *Log) metaWithError(err error) Meta {
	meta := Meta{"error": err.Error()}
	if err, ok := err.(stackTracer); ok {
		stack := []string{}
		for _, f := range err.StackTrace() {

			t, _ := f.MarshalText()
			stack = append(stack, string(t))
		}

		meta["stack"] = stack
	}

	return meta
}
