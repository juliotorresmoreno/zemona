package middlewares

import (
	"log"
	"net/http"
	"time"

	"github.com/fatih/color"
)

type status struct {
	code int
}

type writterHttp struct {
	status *status
	http.ResponseWriter
}

func (w writterHttp) WriteHeader(status int) {
	w.status.code = status
	w.ResponseWriter.WriteHeader(status)
}

func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writter := writterHttp{&status{}, w}

		h.ServeHTTP(writter, r)

		end := time.Now()
		t := end.Sub(start)

		logWritter := log.Writer()

		if writter.status.code >= 200 && writter.status.code <= 299 {
			d := color.New(color.FgWhite)
			d.Fprintln(logWritter, writter.status.code, r.Method, r.RequestURI, t)
		} else if writter.status.code >= 300 && writter.status.code <= 399 {
			d := color.New(color.FgGreen)
			d.Fprintln(logWritter, writter.status.code, r.Method, r.RequestURI, t)
		} else if writter.status.code >= 100 && writter.status.code <= 199 {
			d := color.New(color.FgCyan)
			d.Fprintln(logWritter, writter.status.code, r.Method, r.RequestURI, t)
		} else {
			d := color.New(color.FgRed)
			d.Fprintln(logWritter, writter.status.code, r.Method, r.RequestURI, t)
		}
	})
}
