// smina_post
// Writes all docking scores into one file `vs_scores.txt`
// Result files with scores <= -8.5 are put into the current dir
// parallel version with max. 5 goroutines - fastest with no ReadFile errors.
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/apahl/vstools/internal/fileutls"
	"github.com/korovkin/limiter"

	"github.com/apahl/utls"
)

// Score holds Ligand Id and the score value
type Score struct {
	ID    string
	Value float32
}

const (
	minScore      = -8.5
	numGoroutines = 5
)

func getDir() string {
	if len(os.Args) < 2 {
		utls.QuitOnError(errors.New("missing directory"))
	}
	result := os.Args[1]
	return result
}

func getScore(scoreDir, logFile string) Score {
	ligID := strings.TrimSuffix(logFile, ".log")
	logContent, err := ioutil.ReadFile(filepath.Join(scoreDir, logFile))
	if err != nil {
		return Score{ligID, 0.0}
	}
	logText := string(logContent)
	start := strings.Index(logText, "1    ")
	if start < 0 {
		return Score{ligID, 0.0}
	}
	stop := strings.Index(logText[start:], "\n")
	if stop < 0 {
		return Score{ligID, 0.0}
	}
	fields := strings.Fields(logText[start : start+stop])
	if len(fields) < 2 {
		return Score{ligID, 0.0}
	}
	val64, err := strconv.ParseFloat(fields[1], 32)
	utls.QuitOnError(err)
	value := float32(val64)
	return Score{ligID, value}
}

func copyLigand(ligID, srcDir, dstDir string) {
	for _, ext := range []string{".pdbqt", ".log", ".terms"} {
		src := filepath.Join(srcDir, fmt.Sprintf("%s%s", ligID, ext))
		dst := filepath.Join(dstDir, fmt.Sprintf("%s%s", ligID, ext))
		err := fileutls.Copy(src, dst)
		if ext != ".terms" {
			// ".terms" may not exist, so we don't error out
			// when it doesn't
			utls.QuitOnError(err)
		}
	}
}

// scanScores scans the Smina logfiles
func scanScores(scoreDir string) {
	files, err := ioutil.ReadDir(scoreDir)
	utls.QuitOnError(err)
	limit := limiter.NewConcurrencyLimiter(numGoroutines)
	fmt.Println("")
	for _, file := range files {
		if logFile := file.Name(); strings.HasSuffix(logFile, ".log") {
			limit.Execute(func() {
				score := getScore(scoreDir, logFile)
				fmt.Printf("%s\t%.1f\n", score.ID, score.Value)
				if score.Value <= minScore {
					copyLigand(score.ID, scoreDir, "./")
				}
			})
		}
	}
	limit.Wait()
}

func main() {
	scoreDir := getDir()
	scanScores(scoreDir)
}
