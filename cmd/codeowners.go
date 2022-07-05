/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"github.com/dicksontung/codeowners"
	"github.com/spf13/cobra"
)

// codeownersCmd represents the codeowners command
var codeownersCmd = &cobra.Command{
	Use:   "codeowners",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(files) == 0 {
			files = append(files, ".")
		}
		ruleset, err := loadCodeowners(codeownersFile)
		if err != nil {
			return err
		}
		for _, file := range files {
			rule, err := ruleset.Match(file)
			if err != nil {
				return err
			}
			if rule != nil {
				fmt.Printf("%v %v\n", file, rule.Owners)
			}
		}

		return nil
	},
}

func loadCodeowners(path string) (codeowners.Ruleset, error) {
	if path == "" {
		return codeowners.LoadFileFromStandardLocation()
	}
	return codeowners.LoadFile(path)
}

var (
	codeownersFile string
	files          []string
)

func init() {
	codeownersCmd.PersistentFlags().StringVarP(&codeownersFile, "codeownersfile", "", "", "CODEOWNERS file path")
	codeownersCmd.PersistentFlags().StringSliceVarP(&files, "files", "f", make([]string, 0), "files to check ownership")
	rootCmd.AddCommand(codeownersCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// codeownersCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// codeownersCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
