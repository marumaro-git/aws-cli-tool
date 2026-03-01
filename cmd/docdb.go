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

var updateSingleCmd = &cobra.Command{
	Use:   "update-single",
	Short: "Single event update with Last Write Wins",
	Long:  `Demonstrate single event update: newer events are applied, older events are rejected.`,
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

		logger.Info(ctx, "Running single event update demo (Last Write Wins)...")
		if err := docdbUseCase.UpdateUserSingle(ctx); err != nil {
			logger.Error(ctx, err)
			os.Exit(1)
		}
	},
}

var updateBatchCmd = &cobra.Command{
	Use:   "update-batch",
	Short: "Batch event update with time-ordered processing",
	Long:  `Demonstrate batch event update: events are sorted by event_time and applied in order within a transaction.`,
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

		logger.Info(ctx, "Running batch event update demo...")
		if err := docdbUseCase.UpdateUserBatch(ctx); err != nil {
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
	docdbCmd.AddCommand(updateSingleCmd)
	docdbCmd.AddCommand(updateBatchCmd)
}
