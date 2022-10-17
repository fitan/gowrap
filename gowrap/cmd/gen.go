/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/fitan/gowrap/generator"
	"github.com/fitan/gowrap/pkg"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := pkg.Load(pkgDir, true)
		if err != nil {
			log.Fatalf("load package error: %v", err)
		}

		gen, err := generator.NewGen(generator.GenOption{Pkg: p})
		if err != nil {
			log.Fatalf("new generator error: %v", err)
		}

		for _, v := range genTo {
			ss := strings.Split(v, ":")
			if len(ss) != 2 {
				log.Fatalf("invalid gen to: %s", v)
			}

			templateName, outputDir := ss[0], ss[1]

			err = generator.GenToTemplate(templateName, outputDir, gen)
			if err != nil {
				log.Fatalf("gen to %s error: %v", v, err)
				return
			}
		}
	},
}

var genTo []string
var pkgDir string

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.Flags().StringSliceVarP(&genTo, "to", "g", []string{}, "generate to")
	genCmd.Flags().StringVarP(&pkgDir, "pkg dir", "p", "./", "package dir")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
