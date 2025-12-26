/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/marumaro-git/aws-cli-tool/internal/infrastructure/sqs"
	"github.com/marumaro-git/aws-cli-tool/internal/usecase"
	"github.com/spf13/cobra"
)

// sqsCmd represents the sqs command
var sqsCmd = &cobra.Command{
	Use:   "sqs",
	Short: "SQS operations for LocalStack",
	Long: `Perform SQS operations like sending and receiving messages using LocalStack.
	
This command will run the SQS service demo which receives messages from 'recieve-queue' 
and sends a test message to 'send-queue'.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Running SQS service demo...")
		sqsClient := sqs.NewSQSClient(cmd.Context())
		sqsUseCase := usecase.NewMessageUseCase(sqsClient)
		sqsUseCase.TransferMessages(cmd.Context())
		fmt.Println("✅ SQS service demo completed.")
	},
}

func init() {
	rootCmd.AddCommand(sqsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sqsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sqsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
