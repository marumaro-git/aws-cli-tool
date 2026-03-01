package cmd

import (
	"os"
	"time"

	"github.com/marumaro-git/aws-cli-tool/internal/infrastructure/docdb"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/logger"
	"github.com/marumaro-git/aws-cli-tool/internal/usecase"
	"github.com/spf13/cobra"
)

var docdbCmd = &cobra.Command{
	Use:   "docdb",
	Short: "DocumentDB (MongoDB compatible) operations",
	Long:  `Perform DocumentDB operations for eventual consistency verification using MongoDB-compatible API.`,
}

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Test DocumentDB connection",
	Long:  `Connect to DocumentDB (MongoDB compatible) on LocalStack and verify the connection.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := logger.NewSlogLogger()

		logger.Info(ctx, "Connecting to DocumentDB...")
		client, err := docdb.NewDocDBClient(ctx)
		if err != nil {
			logger.Error(ctx, err)
			os.Exit(1)
		}
		defer client.Close(ctx)

		logger.Info(ctx, "Successfully connected to DocumentDB")
	},
}

var insertCmd = &cobra.Command{
	Use:   "insert",
	Short: "Insert sample events",
	Long:  `Insert sample events with timestamp-based IDs into DocumentDB.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := logger.NewSlogLogger()

		logger.Info(ctx, "Connecting to DocumentDB...")
		client, err := docdb.NewDocDBClient(ctx)
		if err != nil {
			logger.Error(ctx, err)
			os.Exit(1)
		}
		defer client.Close(ctx)

		docdbUseCase := usecase.NewDocDBUseCase(client, logger)

		logger.Info(ctx, "Inserting sample events...")
		if err := docdbUseCase.InsertSampleEvents(ctx); err != nil {
			logger.Error(ctx, err)
			os.Exit(1)
		}
		logger.Info(ctx, "Sample events inserted successfully")
	},
}

var consistencyCmd = &cobra.Command{
	Use:   "consistency",
	Short: "Check eventual consistency",
	Long:  `Verify that events are eventually consistent by checking chronological order.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := logger.NewSlogLogger()

		logger.Info(ctx, "Connecting to DocumentDB...")
		client, err := docdb.NewDocDBClient(ctx)
		if err != nil {
			logger.Error(ctx, err)
			os.Exit(1)
		}
		defer client.Close(ctx)

		docdbUseCase := usecase.NewDocDBUseCase(client, logger)

		logger.Info(ctx, "Checking eventual consistency...")
		if err := docdbUseCase.CheckEventualConsistency(ctx, 10*time.Second); err != nil {
			logger.Error(ctx, err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(docdbCmd)
	docdbCmd.AddCommand(connectCmd)
	docdbCmd.AddCommand(insertCmd)
	docdbCmd.AddCommand(consistencyCmd)
}
