/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/marumaro-git/aws-cli-tool/internal/infrastructure/dynamodb"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/customerror"
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
		fmt.Println("🚀 Running DynamoDB service demo...")
		dynamodbClient := dynamodb.NewDynamoDBClient(cmd.Context())
		dynamodbUseCase := usecase.NewDynamoDBUseCase(dynamodbClient)
		dynamodbUseCase.CheckTTLProcess(cmd.Context())
		fmt.Println("✅ DynamoDB service demo completed.")

		fmt.Println("🚀 Running Batch Write Items demo...")
		dynamodbUseCase.BatchWriteItems(cmd.Context())
		fmt.Println("✅ Batch Write Items demo completed.")

		fmt.Println("🚀 Checking ItemNotFound error handling...")
		err := dynamodbUseCase.ItemNotFound(cmd.Context())
		if err != nil {
			errs := customerror.HandleError(err)
			fmt.Printf("Handled Error - StatusCode: %d, Message: %s\n", errs.StatusCode, errs.Message)
			os.Exit(errs.StatusCode)
		} else {
			fmt.Println("❌ ItemNotFound error handling check completed successfully.")
		}
	},
}

var gureguCmd = &cobra.Command{
	Use:   "guregu",
	Short: "Dynamo guregu library specific operations",
	Long:  `Commands related to Dynamo library operations.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Running Dynamo guregu library demo...")
		dynamodbClient := dynamodb.NewExpressiveDynamoDBClient(cmd.Context())
		tableClient := dynamodbClient.NewExpressiveTableClient()
		dynamodbUseCase := usecase.NewDynamoDBExpressiveUseCase(tableClient)
		dynamodbUseCase.CheckTTLProcess(cmd.Context())
		fmt.Println("✅ Dynamo guregu library demo completed.")

		fmt.Println("🚀 Running Batch Write Items demo...")
		dynamodbUseCase.BatchWriteItems(cmd.Context())
		fmt.Println("✅ Batch Write Items demo completed.")
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
