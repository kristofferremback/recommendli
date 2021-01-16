package server

import (
	"encoding/json"
	"net/http"

	"github.com/kristofferostlund/recommendli/pkg/httphelpers"
)

type responseOpts struct {
	status      int
	contentType string
}

type responseOptFunc func(ropts *responseOpts)

func applicationTypeJSON() responseOptFunc {
	return func(ropts *responseOpts) {
		ropts.contentType = "application/json"
	}
}

func status(status int) responseOptFunc {
	return func(ropts *responseOpts) {
		ropts.status = status
	}
}

func (s *Server) respond(w http.ResponseWriter, body []byte, opts ...responseOptFunc) {
	ropts := &responseOpts{
		status:      200,
		contentType: "plain/text",
	}
	for _, opt := range opts {
		opt(ropts)
	}

	w.WriteHeader(ropts.status)
	w.Write(body)
}

func (s *Server) internalServerError(w http.ResponseWriter) {
	httphelpers.InternalServerError(w)
}

func (s *Server) json(w http.ResponseWriter, data interface{}, opts ...responseOptFunc) {
	b, err := json.Marshal(data)
	if err != nil {
		s.log.Error("Failed to marshal data", err)
		httphelpers.InternalServerError(w)
		return
	}

	ropts := []responseOptFunc{applicationTypeJSON()}
	s.respond(w, b, append(ropts, opts...)...)
}
