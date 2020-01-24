package main

import (
	"net/http"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func startRestAPI() {
	router := mux.NewRouter()
	router.Methods("GET").Path("/health").HandlerFunc(getHealth)
	router.Use(errorLogger)
	go func() {
		log.Fatal(http.ListenAndServe(":7070", router))
   }()
}

func getHealth(w http.ResponseWriter, r *http.Request) {
	logger := log.WithField("action", "get health")
	if _, err := w.Write([]byte("All is well")); err != nil {
		logger.WithError(err).Error("could not write HTTP response")
	}
}

// responseInterceptor implements http.ResponseWriter but captures status codes and response text
type responseInterceptor struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
	response    string
}

// WriteHeader passes through and captures statusCode
func (w *responseInterceptor) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.code = statusCode
	w.wroteHeader = true
}

func (w *responseInterceptor) Write(p []byte) (n int, err error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	w.response = w.response + string(p)
	return w.ResponseWriter.Write(p)
}

// errorLogger is a middleware that logs status code and response body for each request that
// returns status code >= 400
func errorLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		interceptor := &responseInterceptor{ResponseWriter: w}
		next.ServeHTTP(interceptor, r)
		if interceptor.code >= 400 {
			log.WithFields(log.Fields{
				"url":           r.URL.String(),
				"status":        interceptor.code,
				"response_text": interceptor.response,
			}).Infof("HTTP error response")
		}
	})
}

