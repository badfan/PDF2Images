package cmd

import (
	"context"
	"log"
	"pdf2images/rpc"
	"pdf2images/services"
	"pdf2images/storage"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"google.golang.org/grpc"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:       "start",
	Short:     "Launches pdf2images µ-service",
	Args:      cobra.OnlyValidArgs,
	ValidArgs: []string{"local", "dev", "stg", "prd"},
	Run: func(cmd *cobra.Command, args []string) {
		// Viper
		err := config.NewConfig(config.Options{
			ProjectID:   viper.GetString("firebase_project_id"),
			ServiceName: viper.GetString("firestore_collection"),
			Environment: viper.GetString("environment"),
		})
		if err != nil {
			log.Fatalf("failed to create viper instance for config : %v", err)
		}

		// Logger
		zapLogger, err := logging.NewLogger()
		if err != nil {
			log.Fatalf("failed to create logger : %v", err)
		}
		otelLogger := otelzap.New(zapLogger)
		logger := otelLogger.Sugar()
		defer logger.Sync()

		// Context with span
		ctx, span := tracer.NewSpan(context.Background(), "pdf2images.main", nil)
		defer span.End()

		// Context with timeout
		ctx, cancel := context.WithTimeout(ctx, time.Duration(viper.GetInt64("context_timeout"))*time.Second)
		defer cancel()

		// JAEGER
		tp, err := tracer.InitTracer()
		if err != nil {
			logger.Ctx(ctx).Fatalw("failed to create JAEGER tracer provider", "error", err.Error())
		}
		defer tp.Shutdown(ctx)

		// µ-service's file storage client
		blobFileStorageClient, err := storage.NewBlobFileStorageClient(logger)
		if err != nil {
			logger.Ctx(ctx).Fatalw("failed to connect to blob file storage", "error", err.Error())
		}

		// µ-service's service
		imagesService := services.NewPDF2ImagesService(blobFileStorageClient, logger)

		// µ-service's handler
		rpcHandler := rpc.NewRPCHandler(imagesService, logger)

		// RPC server
		rpcServer, listener, err := webserver.NewGRPCServer()
		if err != nil {
			logger.Ctx(ctx).Fatalw("failed to up grpc server", "error", err.Error())
		}
		rpc.RegisterPDF2ImagesServiceServer(rpcServer, rpcHandler)
		go func() {
			if err = rpcServer.Serve(listener); err != nil && err != grpc.ErrServerStopped {
				logger.Ctx(ctx).Fatalw("error occurred while running grpc server", "error", err.Error())
			}
		}()

		logger.Ctx(ctx).Infow("server is up")

		webserver.KeepAliveWithSignals(ctx, logger, func() {
			rpcServer.GracefulStop()
		})
	},
}

func init() {
	startCmd.Flags().StringP(
		"environment",
		"e",
		"local",
		"The environment in which the µ-service will run. Accepted values are: local, dev, stg, prd",
	)
	startCmd.Flags().StringP(
		"firebase_project_id",
		"f",
		"",
		"The id of the Firebase project with configs stored in a Firestore DB",
	)
	startCmd.MarkFlagRequired("firebase_project_id")
	startCmd.Flags().StringP(
		"firestore_collection",
		"c",
		"",
		"The name of the Firestore collection where this µ-service's config are stored. Ideally should be the same as the service name",
	)
	startCmd.MarkFlagRequired("firestore_collection")
	// Bind flags to Viper, so they will be available across the app
	viper.BindPFlag("environment", startCmd.Flags().Lookup("environment"))
	viper.BindPFlag("firebase_project_id", startCmd.Flags().Lookup("firebase_project_id"))
	viper.BindPFlag("firestore_collection", startCmd.Flags().Lookup("firestore_collection"))

	rootCmd.AddCommand(startCmd)
}
