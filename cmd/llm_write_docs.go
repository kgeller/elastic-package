// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cmd

import (
	_ "embed"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/elastic/elastic-package/internal/cobraext"
	"github.com/elastic/elastic-package/internal/packages"
	"github.com/elastic/elastic-package/internal/packages/archetype"
)



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

	manifest , err := archetype.GetManifest(pkgRootDir)
	if err != nil {
		return fmt.Errorf("failed to get package manifest: %w", err)
	}

	llmResponse, err := archetype.GenerateContentWithBedrock(manifest.Name)
	if err != nil {
		return fmt.Errorf("failed to generate documentation content from LLM: %w", err)
	}

	if err := archetype.WriteDocumentationFiles(pkgRootDir, llmResponse); err != nil {
		return err
	}

	cmd.Println("Done")
	return nil
}
