/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init called")
	},
}

var initType string
var initName string
var outputDir string

var m map[string]map[string]string = map[string]map[string]string{
	"service": map[string]string{
		"service": "service.go",
		"new":     "new.go",
	},
	"repo": map[string]string{
		"service": "service.go",
	},
}

func init() {
	initCmd.Flags().StringVarP(&initType, "type", "t", "app", "init type")
	initCmd.Flags().StringVarP(&initName, "name", "n", "app", "init name")
	initCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "output dir")
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
