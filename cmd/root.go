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
	"strconv"
	"sync"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

const (
	SortByDefault = iota
	SortByLines
	SortByNum
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tally",
	Short: "Read stdin and return line:count pairs.",
	Long:  `Basically the same as sort | uniq -c | sort .`,
	Run:   runTally,
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
	rootCmd.Flags().BoolP("number", "n", false, "Convert line to number (float64 internally) and sort by that")
	rootCmd.Flags().IntP("min", "m", 0, "minimum number of matches to print a line")
	rootCmd.Flags().Bool("sum", false, "Show sum of count")
	rootCmd.Flags().BoolP("json", "j", false, "Output as JSON")
	rootCmd.Flags().BoolP("text", "t", true, "Output as text")
	rootCmd.MarkFlagsMutuallyExclusive("json", "text")
	rootCmd.MarkFlagsMutuallyExclusive("string", "number")

}

type LineCount struct {
	Line  string `json:"line"`
	Count int    `json:"count"`
}

type LineMap map[string]int

type LineCountWithSum struct {
	LineCount []LineCount `json:"linecount"`
	Sum       int         `json:"sum,omitempty"`
}

func countLines(r io.Reader) LineMap {
	lines := make(LineMap)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			lines[line]++
		}
	}
	return lines
}

func sortWrap(cmd *cobra.Command, lines LineMap) []LineCount {
	var sortedLines []LineCount
	if flag, _ := cmd.Flags().GetBool("string"); flag {
		sortedLines = sortLines(lines, SortByLines)
	} else if flag, _ := cmd.Flags().GetBool("number"); flag {
		sortedLines = sortLines(lines, SortByNum)
	} else {
		sortedLines = sortLines(lines, SortByDefault)
	}

	return sortedLines
}

func sortLines(lines LineMap, sortKind int) []LineCount {

	sorted := make([]LineCount, 0, len(lines))

	for line, count := range lines {
		sorted = append(sorted, LineCount{line, count})
	}

	switch sortKind {
	case SortByLines:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Line < sorted[j].Line
		})
	case SortByNum:
		sort.Slice(sorted, func(i, j int) bool {
			numI, _ := strconv.ParseFloat(sorted[i].Line, 64)
			numJ, _ := strconv.ParseFloat(sorted[j].Line, 64)
			return numI < numJ
		})
	default:
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Count < sorted[j].Count
		})
	}
	return sorted
}

func readInput(args []string) LineMap {
	var wg sync.WaitGroup
	results := make(chan LineMap, runtime.NumCPU())

	lines := make(LineMap)

	// Read from stdin or files
	if len(args) == 0 {
		lines = countLines(os.Stdin)
	} else {
		for _, fname := range args {
			file, err := os.Open(fname)
			if err != nil {
				log.Fatalf("failed to open file %s: %v", fname, err)
			}
			defer file.Close()
			wg.Add(1)
			go func(f io.Reader) {
				defer wg.Done()
				results <- countLines(f)
			}(file)
		}
		wg.Wait()
		close(results)
		for res := range results {
			for k, v := range res {
				lines[k] += v
			}
		}
	}
	return lines

}

func runTally(cmd *cobra.Command, args []string) {
	var sortedLines []LineCount

	lines := readInput(args)

	// Sorting
	sortedLines = sortWrap(cmd, lines)

	// Reverse if needed
	if reverse, _ := cmd.Flags().GetBool("reverse"); reverse {
		slices.Reverse(sortedLines)
	}

	// Filter by min count
	minCount, _ := cmd.Flags().GetInt("min")
	filtered := make([]LineCount, 0, len(sortedLines))
	for _, v := range sortedLines {
		if v.Count >= minCount {
			filtered = append(filtered, v)
		}
	}

	// Calculate sum if needed
	showSum, _ := cmd.Flags().GetBool("sum")
	sum := 0
	if showSum {
		for _, v := range filtered {
			sum += v.Count
		}
	}

	output := LineCountWithSum{LineCount: filtered, Sum: sum}
	jsonOutput, _ := cmd.Flags().GetBool("json")
	textOutput, _ := cmd.Flags().GetBool("text")

	if jsonOutput {
		out, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			log.Fatalf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(out))
	} else if textOutput {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		for _, v := range output.LineCount {
			fmt.Fprintf(w, "%v\t%v\n", v.Count, v.Line)
		}
		if showSum {
			fmt.Fprintf(w, "==========\nSUM:%v\n", output.Sum)
		}
		w.Flush()
	}
}
