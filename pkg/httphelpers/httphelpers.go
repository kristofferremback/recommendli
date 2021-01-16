package httphelpers

import (
	"net/http"
	"time"
)

func InternalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func ClearCookie(w http.ResponseWriter, c *http.Cookie) {
	c.Value = ""
	c.Expires = time.Unix(0, 0)
	http.SetCookie(w, c)
}
