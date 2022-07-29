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
	fileOwners             = map[string]map[string]struct{}{} //map[file]map[owner]bool
	filesApproved          = map[string]bool{}
	approversApproved      = map[string]bool{}
	comment                = Comment{}
	errorIfNotApproved     = false
	disableCodeOwnersCheck = false
	enableWebhook          = false
	webhookSecret          = ""
	webhookUrl             = ""
	gitCli                 *gitlab.Client
)

type AddWebhookOptions struct {
	URL                      string
	ConfidentialNoteEvents   bool
	PushEvents               bool
	IssuesEvents             bool
	ConfidentialIssuesEvents bool
	MergeRequestsEvents      bool
	TagPushEvents            bool
	NoteEvents               bool
	JobEvents                bool
	PipelineEvents           bool
	WikiPageEvents           bool
	EnableSSLVerification    bool
	Token                    string
}

type Comment struct {
	AllApproved   bool            `yaml:"all_approved"`
	FilesApproved map[string]bool `yaml:"files_approved"`
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
		gitCli, _ = gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.GitlabURL+"/api/v4"))
		if enableWebhook {
			addWebhook()
		}
		changes, _, err := gitCli.MergeRequests.GetMergeRequestChanges(config.ProjectID, config.MergeRequestID, nil)
		if err != nil {
			return err
		}
		for _, change := range changes.Changes {
			checkCodeOwner(change.OldPath)
			checkCodeOwner(change.NewPath)
			fileChanges[change.OldPath] = struct{}{}
			fileChanges[change.NewPath] = struct{}{}
		}
		ruleset, err := loadCodeowners(codeownersFile)
		if err != nil {
			return err
		}
		for file, _ := range fileChanges {
			rule, err := ruleset.Match(file)
			if err != nil {
				return err
			}
			if rule != nil {
				if fileOwners[file] == nil {
					fileOwners[file] = map[string]struct{}{}
				}
				for _, o := range rule.Owners {
					fileOwners[file][o.Value] = struct{}{}
				}
			}
		}
		fmt.Printf("Approval needed: %v \n", fileOwners)
		for key := range fileOwners {
			filesApproved[key] = false
		}
		app, _, err := gitCli.MergeRequestApprovals.GetConfiguration(config.ProjectID, config.MergeRequestID)
		if err != nil {
			return err
		}
		for _, a := range app.ApprovedBy {
			approversApproved[a.User.Username] = true
		}
		fmt.Printf("Approval done: %v \n", approversApproved)
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
		if errorIfNotApproved && !allApproved {
			fmt.Fprintln(os.Stderr, "Error: not approved by all owners")
			os.Exit(1)
		}
		return nil
	},
}

func checkCodeOwner(path string) {
	if !disableCodeOwnersCheck && contains(CodeOwnersLocations, path) {
		fmt.Fprintf(os.Stderr, "Error: %v changed \n", path)
		os.Exit(1)
	}
}

func addWebhook() {
	if webhookExist() {
		return
	}
	addWebhookOptions := AddWebhookOptions{
		URL:                      webhookUrl,
		ConfidentialNoteEvents:   false,
		PushEvents:               false,
		IssuesEvents:             false,
		ConfidentialIssuesEvents: false,
		MergeRequestsEvents:      true,
		TagPushEvents:            false,
		NoteEvents:               false,
		JobEvents:                false,
		PipelineEvents:           false,
		WikiPageEvents:           false,
		EnableSSLVerification:    false,
		Token:                    webhookSecret,
	}
	_, _, err := gitCli.Projects.AddProjectHook(config.ProjectID, &gitlab.AddProjectHookOptions{
		EnableSSLVerification: &addWebhookOptions.EnableSSLVerification,
		PushEvents:            &addWebhookOptions.PushEvents,
		MergeRequestsEvents:   &addWebhookOptions.MergeRequestsEvents,
		Token:                 &addWebhookOptions.Token,
		URL:                   &addWebhookOptions.URL,
	})
	if err != nil {
		fmt.Printf("unable to add webhook: %v \n", err)
	}
}

func webhookExist() bool {
	hooks, _, err := gitCli.Projects.ListProjectHooks(config.ProjectID, nil, nil)
	if err != nil {
		fmt.Printf("unable to get webhooks: %v \n", err)
	}
	for _, hook := range hooks {
		if hook.URL == webhookUrl {
			return true
		}
	}
	return false
}

func init() {
	approvalsCmd.Flags().StringVarP(&config.GitlabURL, "server-url", "u", "", "gitlab_url default to "+defaultGitlabUrl)
	approvalsCmd.Flags().StringVarP(&config.ProjectID, "project-id", "p", viper.GetString("CI_PROJECT_ID"), "project id")
	approvalsCmd.Flags().IntVarP(&config.MergeRequestID, "merge-request-iid", "m", viper.GetInt("CI_MERGE_REQUEST_IID"), "merge request id")
	approvalsCmd.Flags().StringVarP(&config.Token, "job-token", "t", viper.GetString("CI_JOB_TOKEN"), "gitlab token")
	approvalsCmd.Flags().StringVarP(&codeownersFile, "codeownersfile", "", "", "CODEOWNERS file path")
	approvalsCmd.Flags().BoolVarP(&errorIfNotApproved, "error", "", false, "error on exit if not approved")
	approvalsCmd.Flags().BoolVarP(&disableCodeOwnersCheck, "disable-codeowners-check", "", false, "disable CODEOWNERS file check")
	approvalsCmd.Flags().BoolVarP(&enableWebhook, "enable-webhook", "", false, "Automatically add webhook if it does not exist")
	approvalsCmd.Flags().StringVarP(&webhookUrl, "webhook-url", "", "", "webhook url to add")
	approvalsCmd.Flags().StringVarP(&webhookSecret, "webhook-token", "", viper.GetString("CI_WEBHOOK_TOKEN"), "webhook token used to validate webhook")
	rootCmd.AddCommand(approvalsCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// approvalsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// approvalsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
