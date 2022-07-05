/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate a config file",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		yamlData, err := yaml.Marshal(&config)
		if err != nil {
			return err
		}
		fmt.Println(string(yamlData))
		return nil
	},
}

type Config struct {
	GitlabURL      string `yaml:"gitlab_url"`
	Token          string `yaml:"token"`
	ProjectID      string `yaml:"project_id"`
	MergeRequestID int    `yaml:"merge_request_id"`
}

const (
	defaultGitlabUrl = "https://gitlab.com"
)

var (
	config Config
)

func init() {
	configCmd.PersistentFlags().StringVarP(&config.GitlabURL, "gitlab_url", "u", defaultGitlabUrl, "gitlab_url default to "+defaultGitlabUrl)
	configCmd.PersistentFlags().StringVarP(&config.ProjectID, "project_id", "p", "", "project id")
	configCmd.PersistentFlags().IntVarP(&config.MergeRequestID, "merge_request_id", "m", 0, "merge request id")
	configCmd.PersistentFlags().StringVarP(&config.Token, "token", "t", "", "gitlab access token")

	configCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
