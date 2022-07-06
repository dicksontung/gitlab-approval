/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"

	"github.com/spf13/cobra"
)

// changesCmd represents the changes command
var changesCmd = &cobra.Command{
	Use:   "changes",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		gitCli, err := gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.GitlabURL+"/api/v4"))
		if err != nil {
			return err
		}
		changes, _, err := gitCli.MergeRequests.GetMergeRequestChanges(config.ProjectID, config.MergeRequestID, nil)
		if err != nil {
			return err
		}
		for _, change := range changes.Changes {
			fileChanges[change.NewPath] = struct{}{}
			fileChanges[change.OldPath] = struct{}{}
		}
		for key, _ := range fileChanges {
			fmt.Println(key)
		}
		return nil
	},
}

var (
	fileChanges = map[string]struct{}{}
)

func init() {
	changesCmd.Flags().StringVarP(&config.GitlabURL, "server-url", "u", "", "gitlab_url default to "+defaultGitlabUrl)
	changesCmd.Flags().StringVarP(&config.ProjectID, "project-id", "p", viper.GetString("CI_PROJECT_ID"), "project id")
	changesCmd.Flags().IntVarP(&config.MergeRequestID, "merge-request-iid", "m", viper.GetInt("CI_MERGE_REQUEST_IID"), "merge request id")
	changesCmd.Flags().StringVarP(&config.Token, "job-token", "t", viper.GetString("CI_JOB_TOKEN"), "gitlab token")
	rootCmd.AddCommand(changesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// changesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// changesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
