/*
Copyright © 2024 Eric Osborne
No header.
*/
package cmd

import (
	"bufio"
	"fmt"
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
		tally(cmd)
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

}

func tally(cmd *cobra.Command) {
	reverse, _ := cmd.Flags().GetBool("descending")

	lines := make(map[string]int)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		lines[line] += 1
	}

	// now sort
	type LineWithCount struct {
		line  string
		count int
	}

	var sortedLines []LineWithCount

	for line, count := range lines {
		sortedLines = append(sortedLines, LineWithCount{line, count})
	}

	// sort

	sort.Slice(sortedLines, func(i, j int) bool {
		return sortedLines[i].count < sortedLines[j].count
	})

	// reverse?
	if reverse {
		slices.Reverse(sortedLines)
	}

	for _, v := range sortedLines {
		fmt.Println(v.line, v.count)
	}
}
