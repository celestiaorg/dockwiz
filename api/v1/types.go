package api

import (
	"net/http"

	"github.com/celestiaorg/dockwiz/pkg/builder"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type RESTApiV1 struct {
	router        *mux.Router
	server        *http.Server
	logger        *zap.Logger
	loggerNoStack *zap.Logger

	productionMode bool
	builder        *builder.Builder
}

type RESTApiV1Options struct {
	ProductionMode bool
	Logger         *zap.Logger
	Builder        *builder.Builder
}
