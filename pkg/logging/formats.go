package logging

import (
	"encoding/json"
	"strings"
)

const ISO8601 = "2006-01-02T15:04:05.000Z"

var _defaultFormatter = FormatJSON

type Formatter func(msg *Message) []byte

func (f Formatter) Apply(log *Log) {
	log.formatter = f
}

func GetFormatterByName(name string) Formatter {
	switch strings.ToLower(name) {
	case "console":
		return FormatConsole
	case "consolepretty":
		fallthrough
	case "console-pretty":
		return FormatConsolePretty
	case "json":
		return FormatJSON
	default:
		return FormatJSON
	}
}

var FormatConsole = Formatter(func(msg *Message) []byte {
	meta, _ := json.Marshal(msg.Meta)

	text := strings.Join([]string{
		msg.Timestamp.UTC().Format(ISO8601),
		msg.Caller.Location(),
		msg.Level.String(),
		msg.Body,
		string(meta),
	}, " ")

	return []byte(text + "\n")
})

var FormatConsolePretty = Formatter(func(msg *Message) []byte {
	meta, _ := json.MarshalIndent(msg.Meta, "", "  ")

	text := strings.Join([]string{
		msg.Timestamp.UTC().Format(ISO8601),
		msg.Caller.Location(),
		msg.Level.String(),
		msg.Body,
		string(meta),
	}, " ")

	return []byte(text + "\n")
})

var FormatJSON = Formatter(func(msg *Message) []byte {
	out := map[string]interface{}{
		"time":    msg.Timestamp.UTC().Format(ISO8601),
		"level":   msg.Level.String(),
		"message": msg.Body,
		"meta":    msg.Meta,
		"caller":  msg.Caller.LocationAbs(),
	}

	b, _ := json.Marshal(out)
	return append(b, []byte("\n")...)
})
