/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"

	"github.com/spf13/cobra"

	"github.com/xanzy/go-gitlab"
)

var (
	changedFiles       []string
	fileOwners         = map[string]map[string]bool{} //map[file]map[owner]bool
	filesApproved      = map[string]bool{}
	approversApproved  = map[string]bool{}
	comment            = Comment{}
	errorIfNotApproved = false
)

type Comment struct {
	AllApproved   bool            `yaml:"allApproved"`
	FilesApproved map[string]bool `yaml:"filesApproved"`
}

// approvalsCmd represents the approvals command
var approvalsCmd = &cobra.Command{
	Use:   "approvals",
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
			changedFiles = append(changedFiles, change.NewPath, change.OldPath)
		}
		if len(changedFiles) == 0 {
			changedFiles = append(changedFiles, ".")
		}
		ruleset, err := loadCodeowners(codeownersFile)
		if err != nil {
			return err
		}
		for _, file := range changedFiles {
			rule, err := ruleset.Match(file)
			if err != nil {
				return err
			}
			if rule != nil {
				if fileOwners[file] == nil {
					fileOwners[file] = map[string]bool{}
				}
				for _, o := range rule.Owners {
					fileOwners[file][o.Value] = true
				}
			}
			filesApproved[file] = false
		}
		fmt.Printf("Approval needed: %v \n", fileOwners)
		app, _, err := gitCli.MergeRequestApprovals.GetConfiguration(config.ProjectID, config.MergeRequestID)
		if err != nil {
			return err
		}
		for _, a := range app.ApprovedBy {
			approversApproved[a.User.Username] = true
		}
		fmt.Printf("approval done: %v \n", approversApproved)
		for file, owners := range fileOwners {
			for o, _ := range owners {
				if approversApproved[o] {
					filesApproved[file] = true
				}
			}
		}
		allApproved := true
		for f, isApproved := range filesApproved {
			if !isApproved {
				fmt.Printf("Approval pending: %v\n", f)
				allApproved = false
			}
		}
		if allApproved {
			fmt.Println("All approval done")
		}
		comment.FilesApproved = filesApproved
		comment.AllApproved = allApproved
		yamlData, err := yaml.Marshal(comment)
		if err != nil {
			return err
		}
		noteBody := "Approval status: \n```\n" + string(yamlData) + "\n```"
		noteOpt := gitlab.CreateMergeRequestNoteOptions{
			Body: &noteBody,
		}
		_, _, err = gitCli.Notes.CreateMergeRequestNote(config.ProjectID, config.MergeRequestID, &noteOpt)
		if err != nil {
			return err
		}
		if errorIfNotApproved {
			fmt.Fprintln(os.Stderr, "Error: not approved by all owners")
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	approvalsCmd.PersistentFlags().StringVarP(&config.GitlabURL, "gitlab_url", "u", viper.GetString("GITLAB_APPROVAL_URL"), "gitlab_url default to "+defaultGitlabUrl)
	approvalsCmd.PersistentFlags().IntVarP(&config.ProjectID, "project_id", "p", viper.GetInt("GITLAB_APPROVAL_PROJECT_ID"), "project id")
	approvalsCmd.PersistentFlags().IntVarP(&config.MergeRequestID, "merge_request_id", "m", viper.GetInt("GITLAB_APPROVAL_MERGE_REQUEST_ID"), "merge request id")
	approvalsCmd.PersistentFlags().StringVarP(&config.Token, "token", "t", viper.GetString("GITLAB_APPROVAL_TOKEN"), "merge request id")
	approvalsCmd.PersistentFlags().StringVarP(&codeownersFile, "codeownersfile", "", "", "CODEOWNERS file path")
	approvalsCmd.PersistentFlags().BoolVarP(&errorIfNotApproved, "error", "", false, "error on exit if not approved")
	rootCmd.AddCommand(approvalsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// approvalsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// approvalsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
