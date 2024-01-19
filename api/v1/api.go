package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func NewRESTApiV1(opts RESTApiV1Options) *RESTApiV1 {
	restAPI := &RESTApiV1{
		router:         mux.NewRouter(),
		logger:         opts.Logger,
		loggerNoStack:  opts.Logger.WithOptions(zap.AddStacktrace(zap.DPanicLevel)),
		productionMode: opts.ProductionMode,
		builder:        opts.Builder,
	}

	restAPI.router.HandleFunc("/", restAPI.IndexPage).Methods(http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodHead)

	restAPI.router.HandleFunc(APIPath.Build(), restAPI.Build).Methods(http.MethodPost)
	restAPI.router.HandleFunc(APIPath.Status(), restAPI.Status).Methods(http.MethodGet)

	return restAPI
}

func (a *RESTApiV1) Serve(addr, originAllowed string) error {
	http.Handle("/", a.router)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-CSRF-Token"})
	originsOk := handlers.AllowedOrigins([]string{originAllowed})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	a.logger.Info(fmt.Sprintf("serving on %s", addr))

	a.server = &http.Server{
		Addr:    addr,
		Handler: handlers.CORS(originsOk, headersOk, methodsOk)(a.router),
	}

	return a.server.ListenAndServe()
}

// Shutdown stops the API server
func (a *RESTApiV1) Shutdown() error {
	if a.server == nil {
		return errors.New("server is not running")
	}
	return a.server.Shutdown(context.Background())
}

func (a *RESTApiV1) GetAllAPIs() []string {
	list := []string{}
	err := a.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		apiPath, err := route.GetPathTemplate()
		if err == nil {
			list = append(list, apiPath)
		}
		return err
	})
	if err != nil {
		a.logger.Error("error while getting all APIs", zap.Error(err))
	}

	return list
}
