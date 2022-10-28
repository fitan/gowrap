/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/fitan/gowrap/generator"
	"github.com/spf13/cobra"
	"log"
	"strings"
	"sync"
	"time"
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
		p, err := generator.Load(pkgDir, true)
		if err != nil {
			log.Fatalf("load package error: %v", err)
		}

		imports, err := generator.LoadMainImports()
		if err != nil {
			log.Fatalf("load main imports error: %v", err)
		}

		option := generator.GenOption{
			Pkg:     p,
			Imports: imports,
		}
		(&option).ExtraImports()

		gen, err := generator.NewGen(option)
		if err != nil {
			log.Fatalf("new generator error: %v", err)
		}

		w := sync.WaitGroup{}

		for _, v := range genTo {
			ss := strings.Split(v, ":")
			if len(ss) != 2 {
				log.Fatalf("invalid gen to: %s", v)
			}

			templateName, outputDir := ss[0], ss[1]


			w.Add(1)
			go func(templateName, outputDir string, gen generator.Gen) {
				defer w.Done()
				t1 := time.Now()
				log.Println("gen to start: ", templateName)
				err = generator.GenToTemplate(templateName, outputDir, gen, process)
				log.Println("gen to end", templateName, "useTime: ", time.Since(t1).String())
				if err != nil {
					log.Fatalf("gen to %s error: %v", v, err)
					return
				}
			} (templateName, outputDir, gen)
		}

		w.Wait()
	},
}

var genTo []string
var pkgDir string
var process bool

func init() {
	rootCmd.AddCommand(genCmd)

	genCmd.Flags().StringSliceVarP(&genTo, "to", "g", []string{}, "generate to")
	genCmd.Flags().StringVarP(&pkgDir, "pkg dir", "d", "./", "package dir")
	genCmd.Flags().BoolVarP(&process, "process", "p", false, "process")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// genCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// genCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
