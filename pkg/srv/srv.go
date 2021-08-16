package srv

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kristofferostlund/recommendli/pkg/logging"
)

type ResponseOpts struct {
	status      int
	contentType string
}

type ResponseOptFunc func(ropts *ResponseOpts)

func ApplicationTypeJSON() ResponseOptFunc {
	return func(ropts *ResponseOpts) {
		ropts.contentType = "application/json"
	}
}

func ApplicationPlainText() ResponseOptFunc {
	return func(ropts *ResponseOpts) {
		ropts.contentType = "plain/text"
	}
}

func Status(status int) ResponseOptFunc {
	return func(ropts *ResponseOpts) {
		ropts.status = status
	}
}

func InternalServerError(w http.ResponseWriter) {
	respond(w, []byte("Internal server error"), Status(http.StatusInternalServerError), ApplicationPlainText())
}

func JSON(w http.ResponseWriter, data interface{}, opts ...ResponseOptFunc) {
	b, err := json.Marshal(data)
	if err != nil {
		logging.GlobaLogger.Error("Failed to marshal data", err)
		InternalServerError(w)
		return
	}

	respond(w, b, append([]ResponseOptFunc{ApplicationTypeJSON()}, opts...)...)
}

func JSONError(w http.ResponseWriter, err error, opts ...ResponseOptFunc) {
	JSON(w, map[string]string{"error": err.Error()}, opts...)
}

func respond(w http.ResponseWriter, body []byte, opts ...ResponseOptFunc) {
	ropts := &ResponseOpts{
		status:      200,
		contentType: "plain/text",
	}
	for _, opt := range opts {
		opt(ropts)
	}

	w.WriteHeader(ropts.status)
	w.Header().Set("content-type", ropts.contentType)
	// nolint errcheck
	w.Write(body)
}

func ClearCookie(w http.ResponseWriter, c *http.Cookie) {
	c.Value = ""
	c.Expires = time.Unix(0, 0)
	http.SetCookie(w, c)
}
