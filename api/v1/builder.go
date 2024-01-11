package api

import (
	"encoding/json"
	"net/http"

	"github.com/celestiaorg/dockwiz/pkg/builder"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// Build is the handler for the /api/v1/build endpoint
func (a *RESTApiV1) Build(resp http.ResponseWriter, req *http.Request) {
	var bOpts builder.BuilderOptions
	if err := json.NewDecoder(req.Body).Decode(&bOpts); err != nil {
		sendJSONError(resp,
			Message{
				Type:    MessageTypeError,
				Slug:    SlugJSONDecodeFailed,
				Title:   "JSON decode failed",
				Message: err.Error(),
			},
			http.StatusBadRequest)
		a.loggerNoStack.Error("JSON decode failed", zap.Error(err))
		return
	}

	res, err := a.builder.AddToBuildQueue(bOpts)
	if err != nil {
		sendJSONError(resp,
			Message{
				Type:    MessageTypeError,
				Slug:    SlugBuildFailed,
				Title:   "build init failed",
				Message: err.Error(),
			},
			http.StatusInternalServerError)
		a.loggerNoStack.Error("build init failed", zap.Error(err))
		return
	}

	if err := sendJSON(resp, res); err != nil {
		a.loggerNoStack.Error("sending JSON response", zap.Error(err))
	}
}

// Status is the handler for the /api/v1/status/{image_id} endpoint
func (a *RESTApiV1) Status(resp http.ResponseWriter, req *http.Request) {
	imageId := mux.Vars(req)["image_id"]

	status, err := a.builder.GetBuildStatus(imageId)
	if err != nil {
		if err == builder.ErrBuildNotFound {
			sendJSONError(resp,
				Message{
					Type:    MessageTypeWarning,
					Slug:    SlugBuildStatusNotFound,
					Title:   "build status not found",
					Message: err.Error(),
				},
				http.StatusNotFound)
			return
		}

		sendJSONError(resp,
			Message{
				Type:    MessageTypeError,
				Slug:    SlugGetBuildStatusFailed,
				Title:   "getting build status failed",
				Message: err.Error(),
			},
			http.StatusInternalServerError)
		a.loggerNoStack.Error("getting build status failed", zap.Error(err))
		return
	}

	// Just to make it more user friendly ;)
	status.StatusString = status.Status.String()

	if err := sendJSON(resp, status); err != nil {
		a.loggerNoStack.Error("sending JSON response", zap.Error(err))
	}
}
