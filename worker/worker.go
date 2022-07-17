package worker

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

//handles results

type Result struct {
	Line    string //line of text where match was found
	LineNum int    //line index
	Path    string //file path where match was found
}

type Results struct {
	Inner []Result
}

func NewResult(line string, lineNum int, path string) Result {
	return Result{line, lineNum, path}
}

func FindInFile(path string, find string) *Results {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	results := Results{make([]Result, 0)}
	scanner := bufio.NewScanner(file)
	lineNum := 1
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), find) {
			r := NewResult(scanner.Text(), lineNum, path)
			results.Inner = append(results.Inner, r)
		}
		lineNum++
	}
	if len(results.Inner) == 0 {
		return nil
	}
	return &results
}
