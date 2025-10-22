package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/cheggaaa/pb/v3"
	flag "github.com/spf13/pflag"
)

const (
	OnlyFile1 = 1
	BothFiles = 2
	OnlyFile2 = 3
)

type Config struct {
	File1    string
	File2    string
	Color    bool
	Stats    bool
	Format   string
	Group    bool
	Workers  int
	Progress bool
}

type ComparisonResult struct {
	Item   string `json:"item"`
	Status int    `json:"status"`
}

type Stats struct {
	OnlyFile1Count int `json:"only_file1_count"`
	BothFilesCount int `json:"both_files_count"`
	OnlyFile2Count int `json:"only_file2_count"`
	TotalCount     int `json:"total_count"`
}

type OutputData struct {
	Results []ComparisonResult `json:"results"`
	Stats   Stats             `json:"stats"`
}

type FileProcessor struct {
	config *Config
	results map[string]int
	mu      sync.RWMutex
	wg      sync.WaitGroup
}

func NewFileProcessor(config *Config) *FileProcessor {
	return &FileProcessor{
		config:  config,
		results: make(map[string]int),
	}
}

func (fp *FileProcessor) processFile(filename string, status int, progress *pb.ProgressBar) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	
	// Count lines in advance for progress display
	if fp.config.Progress {
		tempFile, _ := os.Open(filename)
		tempScanner := bufio.NewScanner(tempFile)
		for tempScanner.Scan() {
			lineCount++
		}
		tempFile.Close()
		
		if progress != nil {
			progress.SetTotal(int64(lineCount))
		}
	}

	file.Seek(0, 0)
	scanner = bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fp.mu.Lock()
		if existingStatus, exists := fp.results[line]; exists {
			if existingStatus == OnlyFile1 && status == OnlyFile2 {
				fp.results[line] = BothFiles
			}
		} else {
			fp.results[line] = status
		}
		fp.mu.Unlock()

		if progress != nil {
			progress.Increment()
		}
	}

	return scanner.Err()
}

func (fp *FileProcessor) ProcessFiles() error {
	var progress1, progress2 *pb.ProgressBar
	
	if fp.config.Progress {
		progress1 = pb.StartNew(0)
		progress1.SetTemplateString(`{{string . "prefix"}} {{bar . "[" "=" ">" "-" "]"}} {{percent .}} {{string . "suffix"}}`)
		progress1.Set("prefix", "Processing file 1...")
	}

	err := fp.processFile(fp.config.File1, OnlyFile1, progress1)
	if err != nil {
		return err
	}

	if progress1 != nil {
		progress1.Finish()
	}

	if fp.config.Progress {
		progress2 = pb.StartNew(0)
		progress2.SetTemplateString(`{{string . "prefix"}} {{bar . "[" "=" ">" "-" "]"}} {{percent .}} {{string . "suffix"}}`)
		progress2.Set("prefix", "Processing file 2...")
	}

	err = fp.processFile(fp.config.File2, OnlyFile2, progress2)
	if err != nil {
		return err
	}

	if progress2 != nil {
		progress2.Finish()
	}

	return nil
}

func (fp *FileProcessor) GetStats() Stats {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	stats := Stats{}
	for _, status := range fp.results {
		switch status {
		case OnlyFile1:
			stats.OnlyFile1Count++
		case BothFiles:
			stats.BothFilesCount++
		case OnlyFile2:
			stats.OnlyFile2Count++
		}
	}
	stats.TotalCount = len(fp.results)
	return stats
}

func (fp *FileProcessor) GetResults() []ComparisonResult {
	fp.mu.RLock()
	defer fp.mu.RUnlock()

	var results []ComparisonResult
	for item, status := range fp.results {
		results = append(results, ComparisonResult{Item: item, Status: status})
	}
	
	// Sort alphabetically
	sort.Slice(results, func(i, j int) bool {
		return results[i].Item < results[j].Item
	})
	
	return results
}

func printResults(results []ComparisonResult, config *Config) {
	if !config.Color {
		color.NoColor = true
	}

	onlyFile1Color := color.New(color.FgBlue, color.Bold)
	bothFilesColor := color.New(color.FgGreen, color.Bold)
	onlyFile2Color := color.New(color.FgRed, color.Bold)

	switch config.Format {
	case "json":
		outputData := OutputData{
			Results: results,
			Stats:   calculateStats(results),
		}
		jsonData, _ := json.MarshalIndent(outputData, "", "  ")
		fmt.Println(string(jsonData))

	case "table":
		if config.Group {
			printGroupedTable(results, onlyFile1Color, bothFilesColor, onlyFile2Color)
		} else {
			printSimpleTable(results, onlyFile1Color, bothFilesColor, onlyFile2Color)
		}

	default: // csv
		if config.Group {
			printGroupedCSV(results, onlyFile1Color, bothFilesColor, onlyFile2Color)
		} else {
			printSimpleCSV(results, onlyFile1Color, bothFilesColor, onlyFile2Color)
		}
	}
}

func printGroupedTable(results []ComparisonResult, onlyFile1Color, bothFilesColor, onlyFile2Color *color.Color) {
	// Only in FILE1
	onlyFile1Color.Println("\n=== Items only in FILE1 ===")
	for _, result := range results {
		if result.Status == OnlyFile1 {
			onlyFile1Color.Printf("  %s\n", result.Item)
		}
	}

	// In both files
	bothFilesColor.Println("\n=== Items in both files ===")
	for _, result := range results {
		if result.Status == BothFiles {
			bothFilesColor.Printf("  %s\n", result.Item)
		}
	}

	// Only in FILE2
	onlyFile2Color.Println("\n=== Items only in FILE2 ===")
	for _, result := range results {
		if result.Status == OnlyFile2 {
			onlyFile2Color.Printf("  %s\n", result.Item)
		}
	}
}

func printSimpleTable(results []ComparisonResult, onlyFile1Color, bothFilesColor, onlyFile2Color *color.Color) {
	fmt.Println("\n=== Comparison Results ===")
	for _, result := range results {
		var statusText string
		var colorFunc *color.Color
		
		switch result.Status {
		case OnlyFile1:
			statusText = "Only in FILE1"
			colorFunc = onlyFile1Color
		case BothFiles:
			statusText = "In both files"
			colorFunc = bothFilesColor
		case OnlyFile2:
			statusText = "Only in FILE2"
			colorFunc = onlyFile2Color
		}
		
		colorFunc.Printf("  %-20s [%s]\n", result.Item, statusText)
	}
}

func printGroupedCSV(results []ComparisonResult, onlyFile1Color, bothFilesColor, onlyFile2Color *color.Color) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Header
	writer.Write([]string{"Item", "Category"})

	// Only in FILE1
	for _, result := range results {
		if result.Status == OnlyFile1 {
			writer.Write([]string{result.Item, "Only in FILE1"})
		}
	}

	// In both files
	for _, result := range results {
		if result.Status == BothFiles {
			writer.Write([]string{result.Item, "In both files"})
		}
	}

	// Only in FILE2
	for _, result := range results {
		if result.Status == OnlyFile2 {
			writer.Write([]string{result.Item, "Only in FILE2"})
		}
	}
}

func printSimpleCSV(results []ComparisonResult, onlyFile1Color, bothFilesColor, onlyFile2Color *color.Color) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	for _, result := range results {
		var status string
		switch result.Status {
		case OnlyFile1:
			status = "1"
		case BothFiles:
			status = "2"
		case OnlyFile2:
			status = "3"
		}
		writer.Write([]string{result.Item, status})
	}
}

func calculateStats(results []ComparisonResult) Stats {
	stats := Stats{}
	for _, result := range results {
		switch result.Status {
		case OnlyFile1:
			stats.OnlyFile1Count++
		case BothFiles:
			stats.BothFilesCount++
		case OnlyFile2:
			stats.OnlyFile2Count++
		}
	}
	stats.TotalCount = len(results)
	return stats
}

func printStats(stats Stats, config *Config) {
	if !config.Stats {
		return
	}

	if !config.Color {
		color.NoColor = true
	}

	statsColor := color.New(color.FgCyan, color.Bold)
	statsColor.Println("\n=== Statistics ===")
	
	fmt.Printf("Only in FILE1: %d items\n", stats.OnlyFile1Count)
	fmt.Printf("In both files: %d items\n", stats.BothFilesCount)
	fmt.Printf("Only in FILE2: %d items\n", stats.OnlyFile2Count)
	fmt.Printf("Total: %d items\n", stats.TotalCount)
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVarP(&config.File1, "file1", "1", "", "First file to compare (required)")
	flag.StringVarP(&config.File2, "file2", "2", "", "Second file to compare (required)")
	flag.BoolVar(&config.Color, "color", true, "Enable color output")
	flag.BoolVar(&config.Stats, "stats", false, "Show statistics")
	flag.StringVarP(&config.Format, "format", "f", "csv", "Output format (csv, table, json)")
	flag.BoolVarP(&config.Group, "group", "g", false, "Group output by category")
	flag.IntVarP(&config.Workers, "workers", "w", runtime.NumCPU(), "Number of workers for parallel processing")
	flag.BoolVarP(&config.Progress, "progress", "p", false, "Show progress")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] -1 FILE1 -2 FILE2\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "This tool compares two files and analyzes the presence of each item.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nOutput formats:\n")
		fmt.Fprintf(os.Stderr, "  csv   - Output in CSV format (default)\n")
		fmt.Fprintf(os.Stderr, "  table - Output in table format\n")
		fmt.Fprintf(os.Stderr, "  json  - Output in JSON format\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -1 list1.txt -2 list2.txt --stats --color\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -1 list1.txt -2 list2.txt --format table --group\n", os.Args[0])
	}

	flag.Parse()

	return config
}

func main() {
	config := parseFlags()

	if config.File1 == "" || config.File2 == "" {
		fmt.Fprintf(os.Stderr, "Error: Both file1 and file2 must be specified\n")
		flag.Usage()
		os.Exit(1)
	}

	// Check if files exist
	if _, err := os.Stat(config.File1); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File1 not found: %s\n", config.File1)
		os.Exit(1)
	}
	if _, err := os.Stat(config.File2); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: File2 not found: %s\n", config.File2)
		os.Exit(1)
	}

	processor := NewFileProcessor(config)

	startTime := time.Now()
	err := processor.ProcessFiles()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	processingTime := time.Since(startTime)

	results := processor.GetResults()
	stats := processor.GetStats()

	printResults(results, config)
	printStats(stats, config)

	if config.Progress {
		fmt.Printf("\nProcessing time: %v\n", processingTime)
	}
}
