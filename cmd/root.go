/*
Copyright © 2024 Eric Osborne
No header.
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
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
	rootCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	rootCmd.Flags().BoolP("text", "t", true, "Output as text")
	rootCmd.MarkFlagsMutuallyExclusive("json", "text")

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
	Line  string `json:"line"`
	Count int    `json:"count"`
}

type LineCountWithSum struct {
	LineCount []LineCount `json:"linecount"`
	Sum       int         `json:"sum,omitempty"`
}

type wordMap map[string]int

func sortLines(lines wordMap, sortByLines bool) []LineCount {

	sortedLines := make([]LineCount, 0, len(lines))

	for line, count := range lines {
		sortedLines = append(sortedLines, LineCount{line, count})
	}

	if sortByLines {
		sort.Slice(sortedLines, func(i, j int) bool {
			return sortedLines[i].Line < sortedLines[j].Line
		})
	} else {
		sort.Slice(sortedLines, func(i, j int) bool {
			return sortedLines[i].Count < sortedLines[j].Count
		})
	}

	return sortedLines
}

func tally(cmd *cobra.Command, args []string) {

	lines := make(wordMap)
	var wg sync.WaitGroup
	var results = make(chan map[string]int, runtime.NumCPU()*2)
	var tokens = make(chan struct{}, len(args)+1)
	// read stdin or take the names of one or more files
	if len(args) == 0 {
		CountSingleFile(os.Stdin, results)
	} else {
		for _, fname := range args {
			// open the file
			file, err := os.Open(fname)
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

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

	// now sort

	sortByString, _ := cmd.Flags().GetBool("string")
	sortedLines := sortLines(lines, sortByString)

	// reverse?
	reverse, _ := cmd.Flags().GetBool("descending")
	if reverse {
		slices.Reverse(sortedLines)
	}

	txtOutput, _ := cmd.Flags().GetBool("text")
	jsonOutput, _ := cmd.Flags().GetBool("json")

	// figure out sum in case I need it
	showsum, _ := cmd.Flags().GetBool("sum")
	csum := 0
	if showsum {
		for _, v := range sortedLines {
			csum += v.Count
		}
	}

	newSortedLines := make([]LineCount, 0, len(sortedLines))
	limit, _ := cmd.Flags().GetInt("min")
	for _, v := range sortedLines {
		if v.Count >= limit {
			newSortedLines = append(newSortedLines, v)
		}
	}
	sortedLines = newSortedLines

	lcws := LineCountWithSum{sortedLines, csum}

	if jsonOutput {
		// need to add sum to sortedLines somehow, not sure.
		out, err := json.MarshalIndent(lcws, "", " ")
		if err != nil {
			panic("TODO fixme")
		}

		fmt.Println(string(out))
	} else if txtOutput {

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		defer w.Flush()
		for _, v := range lcws.LineCount {
			fmt.Fprintf(w, "%v\t%v\n", v.Count, v.Line)
		}
		if showsum {
			fmt.Fprintf(w, "==========\n")
			fmt.Fprintf(w, "SUM:%v\n", lcws.Sum)
		}
	}

}
