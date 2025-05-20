// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cmd

import (
	// "errors"
	"fmt"
	"context"
	_ "embed"
	"log"

	"github.com/spf13/cobra"

	"github.com/elastic/elastic-package/internal/cobraext"
	// "github.com/elastic/elastic-package/internal/packages"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

const AWS_ACCESS_KEY_ID = ""
const AWS_SECRET_ACCESS_KEY = ""
const AWS_REGION = "us-east-1"
const BEDROCK_MODEL_ID = "anthropic.claude-3-5-sonnet-20240620-v1:0"

type BedrockLLM struct {
	client     *bedrockruntime.Client
	modelID    string
	contentKey string
}

type Response struct {
	Overview string `json:"overview"`
	Setup    string `json:"setup"`
}

const llmWriteDocsLongDescription = `Use this command to write documentation for the package using LLM.
The LLM write docs command generates documentation for the package using a large language model (LLM). 
It analyzes the package files and generates human-readable documentation based on the content and structure 
of the package. The generated documentation is saved in the appropriate format and location within the package.
`

func setupLlmWriteDocsCommand() *cobraext.Command {
	cmd := &cobra.Command{
		Use:   "llm-write-docs",
		Short: "Write documentation for the package using LLM",
		Long:  llmWriteDocsLongDescription,
		Args:  cobra.NoArgs,
		RunE:  llmWriteDocsCommandAction,
	}
	cmd.Flags().BoolP(cobraext.FailFastFlagName, "f", false, cobraext.FailFastFlagDescription)

	return cobraext.NewCommand(cmd, cobraext.ContextPackage)
}

func llmWriteDocsCommandAction(cmd *cobra.Command, args []string) error {
	cmd.Println("Write documentation for the package using LLM")
	
	// _, found, err := packages.FindPackageRoot()
	// if err != nil {
	// 	return fmt.Errorf("locating package root failed: %w", err)
	// }
	// if !found {
	// 	return errors.New("package root not found, you can only author documentation in the package context")
	// }

	// callBedrockViaApi()
	callBedrockViaLibrary()
	

	cmd.Println("Done")
	return nil
}


type Prompt struct {
	Prompt string `json:"prompt"`
}

func callBedrockViaLibrary() {
	// Load AWS configuration with the specified region
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AWS_REGION))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	bedrockSvcClient := bedrockruntime.NewFromConfig(cfg)

	llm, err := bedrock.New(
		bedrock.WithModel(bedrock.ModelAnthropicClaudeV3Sonnet),
		bedrock.WithClient(bedrockSvcClient),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Bedrock LLM: %v", err)
	}

	prompt := `You're an expert documentation writer for Elastic integrations. You are creating a new integration for bitwarden with a datastream for logs. Can you reference the documentation and write me the following 1) an overview and 2) setup instructions for the readme following elastic writing guidelines. Guidelines: Overview: The overview section explains what the integration is, defines the third-party product that is providing data, establishes its relationship to the larger ecosystem of Elastic products, and helps the reader understand how it can be used to solve a tangible problem. The overview should answer the following questions: * What is the integration? * What is the third-party product that is providing data? * What can you do with it? * General description * Basic example Setup: This section should include only setup instructions on the vendor side. For example, for the Cisco ASA integration, users need to configure their Cisco device following the steps found in the Cisco documentation. Note When possible, use links to point to third-party documentation for configuring non-Elastic products since workflows may change without notice. Remove markdown formatting from the response. Do not include any newlines in the response. Please return the response in one line JSON format so that overview and setup instructions are separate keys.`

	ctx := context.Background()
	resp, err := llm.GenerateContent(
		ctx,
		[]llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart(prompt),
				},
			},
		},
		llms.WithMaxTokens(512),
		llms.WithTemperature(0.1),
		llms.WithTopP(1.0),
		llms.WithTopK(100),
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	choices := resp.Choices
	if len(choices) < 1 {
		log.Fatal("Empty response from model")
	}

	fmt.Println(choices[0].Content)
}