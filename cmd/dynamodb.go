/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/marumaro-git/aws-cli-tool/internal/infrastructure/dynamodb"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/customerror"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/logger"
	"github.com/marumaro-git/aws-cli-tool/internal/usecase"
	"github.com/spf13/cobra"
)

// dynamodbCmd represents the dynamodb command
var dynamodbCmd = &cobra.Command{
	Use:   "dynamodb",
	Short: "A brief description of your command",
	Long: `Perform various DynamoDB operations using AWS SDK or dynamo library.
		You can choose between 'sdk' and 'dynamo' subcommands to execute specific operations.`,
}

var sdkCmd = &cobra.Command{
	Use:   "sdk",
	Short: "DynamoDB SDK specific operations",
	Long:  `Commands related to DynamoDB SDK operations.`,
	Run: func(cmd *cobra.Command, args []string) {

		ctx := cmd.Context()
		dynamodbClient := dynamodb.NewDynamoDBClient(ctx)
		logger := logger.NewSlogLogger()
		useCase := usecase.NewDynamoDBUseCase(dynamodbClient, logger)

		logger.Info(ctx, "🚀 Running DynamoDB service demo...")
		useCase.CheckTTLProcess(ctx)
		logger.Info(ctx, "✅ DynamoDB service demo completed.")

		logger.Info(ctx, "🚀 Running Batch Write Items demo...")
		useCase.BatchWriteItems(ctx)
		logger.Info(ctx, "✅ Batch Write Items demo completed.")

		logger.Info(ctx, "🚀 Checking ItemNotFound error handling...")
		err := useCase.ItemNotFound(ctx)
		if err != nil {
			errs := customerror.HandleError(err)
			logger.Error(ctx, err)
			os.Exit(errs.StatusCode)
		} else {
			logger.Info(ctx, "❌ ItemNotFound error handling check completed successfully.")
		}
	},
}

var gureguCmd = &cobra.Command{
	Use:   "guregu",
	Short: "Dynamo guregu library specific operations",
	Long:  `Commands related to Dynamo library operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger := logger.NewSlogLogger()

		dynamodbClient := dynamodb.NewExpressiveDynamoDBClient(ctx)
		tableClient := dynamodbClient.NewExpressiveTableClient()
		dynamodbUseCase := usecase.NewDynamoDBExpressiveUseCase(tableClient, logger)

		logger.Info(ctx, "🚀 Running Dynamo guregu library demo...")
		dynamodbUseCase.CheckTTLProcess(ctx)
		logger.Info(ctx, "✅ Dynamo guregu library demo completed.")

		logger.Info(ctx, "🚀 Running Batch Write Items demo...")
		dynamodbUseCase.BatchWriteItems(ctx)
		logger.Info(ctx, "✅ Batch Write Items demo completed.")
	},
}

func init() {
	rootCmd.AddCommand(dynamodbCmd)
	dynamodbCmd.AddCommand(sdkCmd)
	dynamodbCmd.AddCommand(gureguCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dynamodbCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dynamodbCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
