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
	"sync"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tally",
	Short: "Read stdin and return line:count pairs.",
	Long:  `Basically the same as sort | uniq -c | sort .`,
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
	rootCmd.Flags().BoolP("reverse", "r", false, "Sort in reverse (descending count)")
	rootCmd.Flags().BoolP("string", "s", false, "Sort by string, not count")
	rootCmd.Flags().IntP("min", "m", 0, "minimum number of matches to print a line")
	rootCmd.Flags().BoolP("sum", "", false, "Show sum of count")

}

// countSingleFile returns map[string]int of a single reader.
// TODO: need some tests for this.
func CountSingleFile(r io.Reader, ch chan<- map[string]int) {
	lines := make(map[string]int)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			lines[line] += 1
		}
	}
	ch <- lines
}

type LineCount struct {
	line  string
	count int
}

func tally(cmd *cobra.Command, args []string) {

	lines := make(map[string]int)
	var wg sync.WaitGroup
	var results = make(chan map[string]int, len(args)+1)
	var tokens = make(chan struct{}, len(args)+1)
	// read stdin or take the names of one or more files
	if len(args) == 0 {
		wg.Add(1)
		CountSingleFile(os.Stdin, results)
		wg.Done()
	} else {
		for _, fname := range args {
			// open the file
			file, err := os.Open(fname)
			if err != nil {
				log.Fatal(err)
			}
			wg.Add(1)
			go func(r io.Reader, ch chan<- map[string]int) {
				defer wg.Done()
				tokens <- struct{}{}
				//fmt.Println("in goro with ", fname)
				CountSingleFile(r, ch)
				<-tokens
			}(file, results)
		}
	}
	wg.Wait()
	close(results)
	// now aggregate
	for res := range results {
		for k, v := range res {
			lines[k] += v
		}
	}

	wg.Wait()

	// now sort

	sortedLines := make([]LineCount, 0, len(lines))

	for line, count := range lines {
		sortedLines = append(sortedLines, LineCount{line, count})
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

	// TODO: flag to set a min threshold to display a count
	var csum int
	showsum, _ := cmd.Flags().GetBool("sum")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)

	for _, v := range sortedLines {
		limit, _ := cmd.Flags().GetInt("min")
		csum += v.count
		if v.count >= limit {
			fmt.Fprintf(w, "%v\t%v\n", v.count, v.line)
		}

	}
	if showsum {
		fmt.Fprintf(w, "==========\n")
		fmt.Fprintf(w, "SUM:%v\n", csum)
	}
}
