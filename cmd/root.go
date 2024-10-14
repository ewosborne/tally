/*
Copyright © 2024 Eric Osborne
No header.
*/
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sort"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tally",
	Short: "Read stdin and return line:count pairs.",
	Long:  `Basically the same as sort | uniq -c | sort .`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		tally(cmd, args)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tally.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("descending", "d", false, "Sort descending")
	rootCmd.Flags().BoolP("string", "s", false, "Sort by string, not count")

}

// countSingleFile returns map[string]int of a single reader.
func countSingleFile(r io.Reader, lines map[string]int) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		lines[line] += 1
	}
}

func tally(cmd *cobra.Command, args []string) {

	lines := make(map[string]int)

	// read stdin or take the names of one or more files
	if len(args) == 0 {
		countSingleFile(os.Stdin, lines)
	} else {
		for _, fname := range args {
			file, err := os.Open(fname)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()
			countSingleFile(file, lines)
		}
	}

	// now sort
	type LineWithCount struct {
		line  string
		count int
	}

	sortedLines := make([]LineWithCount, 0, len(lines))

	for line, count := range lines {
		sortedLines = append(sortedLines, LineWithCount{line, count})
	}

	str_sort, _ := cmd.Flags().GetBool("string")
	if str_sort {
		sort.Slice(sortedLines, func(i, j int) bool {
			return sortedLines[i].line < sortedLines[j].line
		})
	} else {
		sort.Slice(sortedLines, func(i, j int) bool {
			return sortedLines[i].count < sortedLines[j].count
		})
	}

	// reverse?
	reverse, _ := cmd.Flags().GetBool("descending")
	if reverse {
		slices.Reverse(sortedLines)
	}

	for _, v := range sortedLines {
		fmt.Println(v.line, v.count)
	}
}
