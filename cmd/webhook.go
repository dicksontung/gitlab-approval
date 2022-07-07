/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

// webhookCmd represents the webhook command
var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Start a server to receive a webhook that will retry all the jobsToRetry in the specified stage",
	Long: `Start a server to receive a webhook that will retry all the jobsToRetry in the specified stage.
---
stages:
  - approval`,
	Run: func(cmd *cobra.Command, args []string) {
		gitCli, _ = gitlab.NewClient(config.Token, gitlab.WithBaseURL(config.GitlabURL+"/api/v4"))
		if len(jobsToRetry) == 0 {
			jobsToRetry = append(jobsToRetry, "approval")
		}
		for _, job := range jobsToRetry {
			jobsToRetrySet[job] = true
		}
		http.HandleFunc("/", httpHandleGetRoot)
		http.HandleFunc("/events", httpHandleGetRoot)
		http.HandleFunc("/healthz", httpHandleGetHealthz)
		fmt.Printf("Listening on port :%v", port)
		err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v", err)
		}
	},
}

func httpHandleGetRoot(w http.ResponseWriter, r *http.Request) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleHttpError(http.StatusBadRequest, err, w, r)
		return
	}
	if len([]rune(webhookToken)) > 0 {
		if webhookToken != r.Header.Get("X-Gitlab-Token") {
			handleHttpError(http.StatusBadRequest, fmt.Errorf("webhook token does not match"), w, r)
			return
		}
	}
	event, err := gitlab.ParseWebhook(gitlab.HookEventType(r), payload)
	if err != nil {
		handleHttpError(http.StatusBadRequest, err, w, r)
		return
	}
	switch event := event.(type) {
	case *gitlab.MergeEvent:
		switch event.ObjectAttributes.Action {
		case "approved", "unapproved", "approval", "unapproval":
			retryJobs(event, w, r)
		default:
			logger.Info("ignoring unsupported event object attribute actions",
				zap.String("event", string(payload)),
			)
		}
	default:
		logger.Info("ignoring unsupported event type",
			zap.String("event", string(payload)),
		)
	}
}

func httpHandleGetHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	return
}

func handleHttpError(code int, err error, w http.ResponseWriter, r *http.Request) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	logger.Error("unable to read payload from request",
		zap.Error(err),
	)
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["error"] = fmt.Sprintf("%v", err)
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func retryJobs(event *gitlab.MergeEvent, w http.ResponseWriter, r *http.Request) {
	mr, _, err := gitCli.MergeRequests.GetMergeRequest(event.Project.ID, event.ObjectAttributes.IID, nil)
	if err != nil {
		handleHttpError(http.StatusInternalServerError, err, w, r)
		return
	}
	if mr.State != "opened" {
		handleHttpError(http.StatusBadRequest, fmt.Errorf("merge request not open"), w, r)
		return
	}
	jobs, _, err := gitCli.Jobs.ListPipelineJobs(event.Project.ID, mr.HeadPipeline.ID, nil)
	if err != nil {
		handleHttpError(500, err, w, r)
		return
	}
	successfulRetried := []string{}
	for _, job := range jobs {
		if jobsToRetrySet[job.Name] {
			_, _, err := gitCli.Jobs.RetryJob(event.Project.ID, job.ID, nil)
			if err != nil {
				handleHttpError(http.StatusInternalServerError, err, w, r)
				return
			}
			successfulRetried = append(successfulRetried, job.Name)
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["jobs_retried"] = fmt.Sprintf("%v", successfulRetried)
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)

	}
}

var (
	webhookToken   = ""
	port           = 80
	jobsToRetry    = []string{}
	jobsToRetrySet = map[string]bool{}
)

func init() {
	webhookCmd.Flags().StringVarP(&config.GitlabURL, "server-url", "u", viper.GetString("CI_SERVER_URL"), "gitlab_url default to "+defaultGitlabUrl)
	webhookCmd.Flags().StringVarP(&config.Token, "job-token", "t", viper.GetString("CI_JOB_TOKEN"), "gitlab token")
	webhookCmd.Flags().StringSliceVarP(&jobsToRetry, "jobs", "", viper.GetStringSlice("CI_JOBS"), "jobs to retry, defaults to 'approval'")
	webhookCmd.Flags().IntVarP(&port, "port", "", viper.GetInt("CI_PORT"), "port number to listen on defaults to 80")
	webhookCmd.Flags().StringVarP(&webhookToken, "webhook-token", "", viper.GetString("CI_WEBHOOK_TOKEN"), "webhook token used to validate payhook")
	rootCmd.AddCommand(webhookCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// webhookCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// webhookCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
