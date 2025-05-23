// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cmd

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/elastic/elastic-package/internal/cobraext"
	"github.com/elastic/elastic-package/internal/packages"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/bedrock"
)

const AWS_REGION = "us-east-1"
const BEDROCK_MODEL_ID = "anthropic.claude-3-5-sonnet-20240620-v1:0"

type BedrockLLM struct {
	client     *bedrockruntime.Client
	modelID    string
	contentKey string
}

type Prompt struct {
	Prompt string `json:"prompt"`
}

type Response struct {
	Overview string `json:"overview"`
	Setup    string `json:"setup"`
}

type ResponseSentences struct {
	Overview []Sentence `json:"overview"`
	Setup    []Sentence `json:"setup"`
}


type Sentence struct {
	Order    int    `json:"order"`
	Content  string `json:"content"`
	Type     string `json:"type"`
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

	// In the console, run "aws-mfa --profile=elastic-siem" first.
	// Credentials will be stored locally and loaded automatically in the context.

	pkgRootDir, found, err := packages.FindPackageRoot()
	if err != nil {
		return fmt.Errorf("locating package root failed: %w", err)
	}
	if !found {
		return errors.New("package root not found, you can only author documentation in the package context")
	}

	manifest , err := getManifest(pkgRootDir)
	if err != nil {
		return fmt.Errorf("failed to get package manifest: %w", err)
	}

	llmResponse, err := generateContentWithBedrock(manifest.Name)
	if err != nil {
		return fmt.Errorf("failed to generate documentation content from LLM: %w", err)
	}

	if err := writeDocumentationFiles(pkgRootDir, llmResponse, cmd); err != nil {
		return err
	}

	cmd.Println("Done")
	return nil
}

// getManifest reads the package manifest from the package root directory.
// It is used to fetch the integration name and send it to the LLM for documentation generation.
func getManifest(pkgRootDir string) (*packages.PackageManifest, error) {
	packageRoot, found, err := packages.FindPackageRoot()
	if err != nil {
		return nil, fmt.Errorf("locating package root failed: %w", err)
	}
	if !found {
		return nil, errors.New("package root not found, you can only create new data stream in the package context")
	}

	manifest, err := packages.ReadPackageManifestFromPackageRoot(packageRoot)
	if err != nil {
		return nil, fmt.Errorf("reading package manifest failed (path: %s): %w", packageRoot, err)
	}

	return manifest, nil
}

// markdownForSentence formats a sentence based on its type for Markdown rendering.
func markdownForSentence(s Sentence) string {
    switch s.Type {
    case "bold":
        return fmt.Sprintf("\n**%s**\n", s.Content)
    case "paragraph":
        return fmt.Sprintf("\n%s\n", s.Content)
    case "list_item":
        return fmt.Sprintf("- %s\n", s.Content)
    case "code_block":
        return fmt.Sprintf("```\n%s\n```\n\n", s.Content)
    default:
        return fmt.Sprintf("%s\n\n", s.Content)
    }
}

// do we really need those generated files? useful for tracking 
// this command could be called at part of the "create" command + stand-alone command

func writeDocumentationFiles(pkgRootDir string, llmResponseSentences ResponseSentences, cmd *cobra.Command) error {
	// Define path and create directories for documentation files
	outputFileDir := filepath.Join(pkgRootDir, "_dev", "build", "docs")
	if err := os.MkdirAll(outputFileDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", outputFileDir, err)
	}

	timestampComment := fmt.Sprintf("\n\n<!-- Generated on: %s -->", time.Now().Format(time.RFC3339))

	llmResponse := transformToMarkdown(llmResponseSentences)

	// Replace literal \\n with actual newlines for proper markdown rendering
	processedOverview := strings.ReplaceAll(llmResponse.Overview, "\\n", "\n")
	overviewContent := processedOverview + timestampComment
	processedSetup := strings.ReplaceAll(llmResponse.Setup, "\\n", "\n")
	setupContent := processedSetup + timestampComment

	overviewFilePath := filepath.Join(outputFileDir, "sections", "generated_overview.md")
	if err := os.WriteFile(overviewFilePath, []byte(overviewContent), 0644); err != nil {
		return fmt.Errorf("failed to write overview to %s: %w", overviewFilePath, err)
	}
	cmd.Printf("Overview successfully written to %s\n", overviewFilePath)
	
	setupFilePath := filepath.Join(outputFileDir, "sections", "generated_setup.md")
	if err := os.WriteFile(setupFilePath, []byte(setupContent), 0644); err != nil {
		return fmt.Errorf("failed to write setup to %s: %w", setupFilePath, err)
	}
	cmd.Printf("Setup successfully written to %s\n", setupFilePath)

	return nil
}

// transformToMarkdown converts the LLM response sentences to a string with markdown format.
func transformToMarkdown(llmResponseSentences ResponseSentences ) Response {
	var llmResponse Response
    var sb strings.Builder

    for _, s := range llmResponseSentences.Overview {
        sb.WriteString(markdownForSentence(s))
    }
	llmResponse.Overview = sb.String()
	sb.Reset()

    for _, s := range llmResponseSentences.Setup {
        sb.WriteString(markdownForSentence(s))
    }
	llmResponse.Setup = sb.String()
	

	return llmResponse
}

func generateContentWithBedrock(packageName string) (ResponseSentences, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile("elastic-siem"), config.WithRegion(AWS_REGION),
	)
	if err != nil {
		return ResponseSentences{}, fmt.Errorf("unable to load SDK config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)
	out, err := stsClient.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if err != nil {
		return ResponseSentences{}, fmt.Errorf("failed to get caller identity: %w", err)
	}

	fmt.Printf("Logged in as: %s\n", *out.Arn)

	bedrockSvcClient := bedrockruntime.NewFromConfig(cfg)

	llm, err := bedrock.New(
		bedrock.WithModel(bedrock.ModelAnthropicClaudeV3Sonnet),
		bedrock.WithClient(bedrockSvcClient),
	)
	if err != nil {
		return ResponseSentences{}, fmt.Errorf("failed to initialize Bedrock LLM: %w", err)
	}

	prompt := fmt.Sprintf(
	`You're an expert documentation writer for Elastic integrations. You are creating a new 
	integration for %s with a datastream for logs. Can you reference the documentation 
	and write me the following 1) an overview and 2) setup instructions for the readme following 
	elastic writing guidelines. 
	Guidelines: 
	Overview: The overview section explains what the integration is, defines the third-party 
	product that is providing data, establishes its relationship to the larger ecosystem of 
	Elastic products, and helps the reader understand how it can be used to solve a tangible 
	problem. The overview should answer the following questions: 
	* What is the integration? 
	* What is the third-party product that is providing data? 
	* What can you do with it? 
	* General description 
	* Basic example 
	Setup: 
	This section should include only setup instructions on the vendor side. For example, for 
	the Cisco ASA integration, users need to configure their Cisco device following the steps 
	found in the Cisco documentation. Note When possible, use links to point to third-party 
	documentation for configuring non-Elastic products since workflows may change without notice. 

	Please return the response in one line JSON format so that overview and setup instructions are separate keys.
	I want to be able to display the overview and setup instructions with markdown formatting. For each piece of text that should be displayed differently from its predecessor: 
	- add an entry to the json with the order of the sentence withe a key named order. 
	- add a key named content with the content of the sentence. 
	- add a key named type with the type of markdown to use when displaying. The exhaustive list of types is: bold, paragraph, list_item, code_block.

	example of the json response: 
	{"overview": [{"order": 1, "content": "The Bitwarden integration enables you to collect and analyze event logs from your Bitwarden organization, providing visibility into password manager usage, security events, and administrative activities across your enterprise.", "type": "paragraph"}, {"order": 2, "content": "Bitwarden is a comprehensive password management solution that helps organizations secure credentials, store sensitive information, and enforce password policies for teams and individuals.", "type": "paragraph"}, {"order": 3, "content": "This integration captures audit logs and security events from Bitwarden's API, allowing you to monitor user authentication, vault access, policy violations, and administrative changes within your password management infrastructure.", "type": "paragraph"}, {"order": 4, "content": "By ingesting Bitwarden logs into Elastic, you can:", "type": "paragraph"}, {"order": 5, "content": "Correlate password manager events with other security data", "type": "list_item"}, {"order": 6, "content": "Detect suspicious authentication patterns", "type": "list_item"}, {"order": 7, "content": "Track compliance with password policies", "type": "list_item"}, {"order": 8, "content": "Gain insights into credential management practices across your organization", "type": "list_item"}, {"order": 9, "content": "The integration supports monitoring activities such as:", "type": "paragraph"}, {"order": 10, "content": "User logins", "type": "list_item"}, {"order": 11, "content": "Vault item access", "type": "list_item"}, {"order": 12, "content": "Sharing events", "type": "list_item"}, {"order": 13, "content": "Policy enforcement", "type": "list_item"}, {"order": 14, "content": "Administrative actions", "type": "list_item"}, {"order": 15, "content": "Example use cases:", "type": "bold"}, {"order": 16, "content": "You can identify users who frequently access shared credentials, detect unusual login patterns that might indicate compromised accounts, or track compliance with password rotation policies by analyzing vault modification events alongside your broader security monitoring strategy.", "type": "paragraph"}], "setup": [{"order": 1, "content": "To configure Bitwarden for log collection, you need to set up API access and enable event logging in your Bitwarden organization.", "type": "paragraph"}, {"order": 2, "content": "Prerequisites:", "type": "bold"}]}
	`, packageName)

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
		llms.WithMaxTokens(1500),
		llms.WithTemperature(0.1),
		llms.WithTopP(1.0),
		llms.WithTopK(100),
	)
	if err != nil {
		return ResponseSentences{}, fmt.Errorf("failed to generate content: %w", err)
	}

	choices := resp.Choices
	if len(choices) < 1 {
		return ResponseSentences{}, errors.New("empty response from model")
	}

	var llmResp ResponseSentences
	err = json.Unmarshal([]byte(choices[0].Content), &llmResp)
	if err != nil {
		return ResponseSentences{}, fmt.Errorf("failed to unmarshal LLM response: %w. Content: %s", err, choices[0].Content)
	}

	return llmResp, nil
}