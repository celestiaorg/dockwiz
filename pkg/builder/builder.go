package builder

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/GoogleContainerTools/kaniko/pkg/buildcontext"
	"github.com/GoogleContainerTools/kaniko/pkg/config"
	"github.com/celestiaorg/dockwiz/pkg/redisqueue"
	"github.com/containerd/containerd/platforms"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

func NewBuilder(redisClient *redis.Client, logger *zap.Logger) *Builder {
	return &Builder{
		redisClient: redisClient,
		logger:      logger,
		Queue:       redisqueue.NewQueue(redisClient, "build_queue"),
		kaniko:      &Kaniko{},
	}
}

func (b *Builder) AddToBuildQueue(opts BuilderOptions) (BuildResult, error) {
	// preparation
	if opts.Image.Name == "" {
		opts.Image.Name = opts.Image.Prefix + uuid.New().String()
	}

	if opts.Image.Tag == "" {
		opts.Image.Tag = defaultImageTag
	}

	if opts.Image.Destination == "" {
		opts.Image.Destination = defaultImageDestination
	}

	if opts.DockerfilePath == "" {
		opts.DockerfilePath = defaultDockerfilePath
	}

	if opts.Git.Branch == "" {
		opts.Git.Branch = defaultGitBranch
	}

	if opts.Git.URL == "" {
		return BuildResult{}, errors.New("git url is required")
	}

	cleanURL, err := cleanGhURL(opts.Git.URL)
	if err != nil {
		return BuildResult{}, fmt.Errorf("cleaning git url: %w", err)
	}
	opts.Git.URL = cleanURL

	err = b.SetBuildStatus(opts.Image.Name, BuildStatusData{
		Status:    StatusPending,
		StartTime: time.Now().UTC(),
		Logs:      fmt.Sprintf("Building image %s:%s\n", opts.Image.Name, opts.Image.Tag),
	})
	if err != nil {
		return BuildResult{}, fmt.Errorf("setting build status: %w", err)
	}

	if err := b.Queue.Enqueue(opts); err != nil {
		return BuildResult{}, fmt.Errorf("adding build to the queue: %w", err)
	}

	return BuildResult{
		ImageName: opts.Image.Name,
		ImageTag:  opts.Image.Tag,
	}, nil
}

// Start starts the builder worker where it dequeues the build requests from the queue
// and builds the images using Kaniko and pushes them to the registry
func (b *Builder) Start() {
	b.logger.Info("Starting builder")
	ctx, cancel := context.WithCancel(context.Background())
	b.startCancelFunc = cancel
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var bOpts BuilderOptions
				if err := b.Queue.Dequeue(&bOpts); err != nil {
					if err == redisqueue.ErrQueueEmpty {
						// Sleep for a random seconds between 1 and 5
						// This is to avoid a thundering herd problem on the redis server
						// https://en.wikipedia.org/wiki/Thundering_herd_problem
						time.Sleep(time.Duration(rand.Intn(5-1)+1) * time.Second)
						continue
					}
					b.logger.Error("dequeue error", zap.Error(err))
					continue
				}

				b.logger.Debug("Got image name from the queue", zap.String("image_name", bOpts.Image.Name))

				bd, err := b.GetBuildStatus(bOpts.Image.Name)
				if err != nil {
					b.logger.Error("getting build status", zap.Error(err))
					continue
				}

				// In a rare case, another instance might be front running the build
				if bd.Status != StatusPending {
					b.logger.Debug("build status is not pending, skipping", zap.String("image_name", bOpts.Image.Name))
					continue
				}

				b.logger.Debug("starting build", zap.String("image_name", bOpts.Image.Name))

				bErr := b.build(bOpts)

				var (
					status  = StatusSucceeded
					bErrMsg = ""
				)
				if bErr != nil {
					status = StatusFailed
					bErrMsg = bErr.Error()
					b.logger.Error("build error:", zap.Error(bErr))
				}

				err = b.UpdateBuildStatus(bOpts.Image.Name, BuildStatusData{
					Status:   status,
					ErrorMsg: bErrMsg,
					EndTime:  time.Now().UTC(),
					Logs:     fmt.Sprintf("Build finished with status %s\n", status.String()),
				})
				if err != nil {
					b.logger.Error("updating build status:", zap.Error(err))
				}
			}
		}
	}()
}

func (b *Builder) Close() error {
	if b.startCancelFunc != nil {
		b.startCancelFunc()
	}

	if b.redisClient == nil {
		return errors.New("redis client is not initialized")
	}
	return b.redisClient.Close()
}

func (b *Builder) build(bOpts BuilderOptions) error {
	// Catch Kaniko logs and write them to the redis
	logsHook := NewCatchLogsHook()
	// Since Kaniko does not receive a logger, we need to add the hook to the global logrus logger
	// Important: Right now using the global logger is the only way to catch the logs
	//   on the other hand, as Kaniko does not support concurrent builds, this should not be a problem
	logrus.AddHook(logsHook)
	logChan, stop := logsHook.StreamNewLogs()
	defer stop()

	go func() {
		for newLogs := range logChan {
			err := b.UpdateBuildStatus(bOpts.Image.Name, BuildStatusData{Logs: newLogs})
			if err != nil {
				b.logger.Error("adding logs to the build status:", zap.Error(err))
			}
		}
	}()

	err := b.UpdateBuildStatus(bOpts.Image.Name, BuildStatusData{Status: StatusBuilding})
	if err != nil {
		return fmt.Errorf("updating build status: %w", err)
	}

	config.BuildContextDir = path.Join(defaultKanikoPath, bOpts.Image.Name)

	// delete the directory if it already exist
	if _, err := os.Stat(config.BuildContextDir); err == nil {
		if err := os.RemoveAll(config.BuildContextDir); err != nil {
			return err
		}
	}

	dockerFilePath, err := filepath.Abs(filepath.Join(config.BuildContextDir, bOpts.DockerfilePath))
	if err != nil {
		return err
	}

	kOpts := &config.KanikoOptions{
		SrcContext: "git://" + bOpts.Git.URL,
		Git: config.KanikoGitOptions{
			Branch:            bOpts.Git.Branch,
			SingleBranch:      bOpts.Git.SingleBranch,
			RecurseSubmodules: bOpts.Git.RecurseSubmodules,
		},
		CustomPlatform: platforms.Format(platforms.Normalize(platforms.DefaultSpec())),
		DockerfilePath: dockerFilePath,
		SnapshotMode:   "full",
		Destinations: []string{
			fmt.Sprintf("%s/%s:%s",
				bOpts.Image.Destination,
				bOpts.Image.Name,
				bOpts.Image.Tag),
		},
		Cache:   true,
		Cleanup: false,
	}

	ctxExec, err := b.kaniko.GetBuildContext(kOpts.SrcContext, buildcontext.BuildOptions{
		GitBranch:            kOpts.Git.Branch,
		GitSingleBranch:      kOpts.Git.SingleBranch,
		GitRecurseSubmodules: kOpts.Git.RecurseSubmodules,
	})
	if err != nil {
		return err
	}

	b.logger.Debug("Getting source context from", zap.String("src_context", kOpts.SrcContext))

	kOpts.SrcContext, err = ctxExec.UnpackTarFromBuildContext()
	if err != nil {
		return err
	}
	b.logger.Debug("Updated source context", zap.String("src_context", kOpts.SrcContext))

	image, err := b.kaniko.DoBuild(kOpts)
	if err != nil {
		return fmt.Errorf("error building image: %w", err)
	}
	if err := b.kaniko.DoPush(image, kOpts); err != nil {
		return fmt.Errorf("error pushing image: %w", err)
	}

	return nil
}

// cleanGhURL removes the scheme from a GitHub URL.
func cleanGhURL(u string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(u, parsedURL.Scheme+"://"), nil
}
