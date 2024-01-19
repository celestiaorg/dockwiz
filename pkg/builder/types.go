package builder

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/celestiaorg/dockwiz/pkg/redisqueue"
	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

const (
	defaultKanikoPath       = "/kaniko"
	defaultDockerfilePath   = "Dockerfile"
	defaultGitBranch        = "main"
	defaultImageTag         = "1h"
	defaultRedisMsgTTL      = 24 * time.Hour
	defaultImageDestination = "ttl.sh"
)

var ErrBuildNotFound = errors.New("build not found")

type Builder struct {
	redisClient     *redis.Client
	logger          *zap.Logger
	Queue           *redisqueue.Queue
	startCancelFunc context.CancelFunc
	kaniko          KanikoInterface
}

type GitOptions struct {
	URL               string `json:"url"`
	Branch            string `json:"branch"`
	SingleBranch      bool   `json:"single_branch"`
	RecurseSubmodules bool   `json:"recurse_submodules"`
}

type ImageOptions struct {
	Prefix string `json:"prefix"`
	Name   string `json:"name"`
	Tag    string `json:"tag"`

	// TODO: add support for multiple destinations with credentials for future
	Destination string `json:"destination"` // Where to push the image
}

type BuilderOptions struct {
	DockerfilePath string       `json:"dockerfile_path"`
	Git            GitOptions   `json:"git_options"`
	CustomPlatform string       `json:"custom_platform"`
	Image          ImageOptions `json:"image"`
	BuildArgs      []string     `json:"build_args"`
}

func (b BuilderOptions) MarshalBinary() ([]byte, error) {
	return json.Marshal(b)
}

func (b *BuilderOptions) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, b)
}

type BuildResult struct {
	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`
}

/*------*/

type BuildStatusData struct {
	Status       BuildStatus `json:"status"`
	StatusString string      `json:"status_string"`
	ErrorMsg     string      `json:"error"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      time.Time   `json:"end_time"`
	Logs         string      `json:"logs"`
}

func (d BuildStatusData) MarshalBinary() ([]byte, error) {
	return json.Marshal(d)
}

func (d *BuildStatusData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, d)
}

type BuildStatus int

const (
	StatusPending BuildStatus = iota + 1
	StatusBuilding
	StatusSucceeded
	StatusFailed
)

func (status BuildStatus) String() string {
	statusStr := [...]string{
		StatusPending:   "pending",
		StatusBuilding:  "building",
		StatusSucceeded: "succeeded",
		StatusFailed:    "failed",
	}

	if status < StatusPending || status > StatusFailed {
		return "unknown"
	}
	return statusStr[status]
}
