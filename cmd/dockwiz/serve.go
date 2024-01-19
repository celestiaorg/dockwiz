package dockwiz

import (
	api "github.com/celestiaorg/dockwiz/api/v1"
	"github.com/celestiaorg/dockwiz/pkg/builder"
	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	flagServeAddr      = "serve-addr"
	flagLogLevel       = "log-level"
	flagProductionMode = "production-mode"

	redisAddr     = "redis-addr"
	redisPassword = "redis-password"
	redisDB       = "redis-db"
)

var flagsServe struct {
	serveAddr      string
	originAllowed  string
	logLevel       string
	productionMode bool

	redisAddr     string
	redisPassword string
	redisDB       int
}

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.PersistentFlags().StringVar(&flagsServe.serveAddr, flagServeAddr, ":9007", "address to serve on")
	serveCmd.PersistentFlags().StringVar(&flagsServe.originAllowed, "origin-allowed", "*", "origin allowed for CORS")

	serveCmd.PersistentFlags().StringVar(&flagsServe.logLevel, flagLogLevel, "info", "log level (e.g. debug, info, warn, error, dpanic, panic, fatal)")
	serveCmd.PersistentFlags().BoolVar(&flagsServe.productionMode, flagProductionMode, false, "production mode (e.g. disable debug logs)")

	serveCmd.PersistentFlags().StringVar(&flagsServe.redisAddr, redisAddr, "localhost:6379", "redis address")
	serveCmd.PersistentFlags().StringVar(&flagsServe.redisPassword, redisPassword, "", "redis password")
	serveCmd.PersistentFlags().IntVar(&flagsServe.redisDB, redisDB, 0, "redis database")
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "serves the Bit Twister API server",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := getLogger(flagsServe.logLevel, flagsServe.productionMode)
		if err != nil {
			return err
		}
		defer func() {
			// The error is ignored because of this issue: https://github.com/uber-go/zap/issues/328
			_ = logger.Sync()
		}()

		// TODO: add a check to make sure not allow users to run this outside a container

		logger.Info("Starting the API server...")

		rdc := redis.NewClient(&redis.Options{
			Addr:     flagsServe.redisAddr,
			Password: flagsServe.redisPassword,
			DB:       flagsServe.redisDB,
		})
		defer func() {
			if err := rdc.Close(); err != nil {
				logger.Fatal("redis client close", zap.Error(err))
			}
		}()
		if _, err := rdc.Ping().Result(); err != nil {
			logger.Fatal("redis ping", zap.Error(err))
			return nil
		}

		opts := api.RESTApiV1Options{
			ProductionMode: flagsServe.productionMode,
			Logger:         logger,
			Builder:        builder.NewBuilder(rdc, logger),
		}

		opts.Builder.Start()

		restAPI := api.NewRESTApiV1(opts)
		defer func() {
			if err := restAPI.Shutdown(); err != nil {
				logger.Fatal("REST API server shutdown", zap.Error(err))
			}
		}()

		if err := restAPI.Serve(flagsServe.serveAddr, flagsServe.originAllowed); err != nil {
			logger.Fatal("REST API server", zap.Error(err))
		}

		return nil
	},
}
