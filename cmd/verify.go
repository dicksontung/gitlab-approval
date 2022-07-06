/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
)

var (
	CodeOwnersLocations = map[string]struct{}{
		"CODEOWNERS":         {},
		".github/CODEOWNERS": {},
		".gitlab/CODEOWNERS": {},
		"docs/CODEOWNERS":    {},
	}
)

func contains(set map[string]struct{}, s string) bool {
	for key, _ := range set {
		if key == s {
			return true
		}
	}
	return false
}

// verifyCmd represents the verify command
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "verify CODEOWNERS file not changed in merge request",
	Long:  `verify CODEOWNERS file not changed in merge request`,
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
			checkCodeOwner(change.OldPath)
			checkCodeOwner(change.NewPath)
		}
		fmt.Println("OK")
		return nil
	},
}

func init() {
	verifyCmd.Flags().StringVarP(&config.GitlabURL, "server-url", "u", "", "gitlab_url default to "+defaultGitlabUrl)
	verifyCmd.Flags().StringVarP(&config.ProjectID, "project-id", "p", viper.GetString("CI_PROJECT_ID"), "project id")
	verifyCmd.Flags().IntVarP(&config.MergeRequestID, "merge-request-iid", "m", viper.GetInt("CI_MERGE_REQUEST_IID"), "merge request id")
	verifyCmd.Flags().StringVarP(&config.Token, "job-token", "t", viper.GetString("CI_JOB_TOKEN"), "gitlab token")
	rootCmd.AddCommand(verifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// verifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// verifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
