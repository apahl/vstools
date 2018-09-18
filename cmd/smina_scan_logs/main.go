// smina_score: parallel version with max. 5 goroutines - fastest with no ReadFile errors.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/korovkin/limiter"

	"github.com/apahl/vstools/internal/calculators"

	"github.com/apahl/utls"
)

// Score is composed of a Compound_Id, its docking score value,
// number of heavy atoms and a status code.
type Score struct {
	id        string
	value     float32
	numHAtoms uint8 // number of heavy atoms
	status    uint8
}

const (
	titleDefault    = "Top Scoring Results from Virtual Screening"
	minScore        = -8.5
	numHAmaxDefault = 255
	ligEffDefault   = 0.0
	numGoroutines   = 5
)

const (
	statusOk uint8 = iota
	statusLowScore
	statusReadErr
	statusStartErr
	statusStopErr
	statusFieldsErr
	statusNumHAErr
	statusImplausErr
)

var (
	flagTopN      int     // TopN values after sorting
	flagMaxValue  float64 // Max value to retrieve, either score or LE
	flagTopNperHA int
	flagNumHAmin  int
	flagNumHAmax  int
	flagSortBy    string
	flagAll       bool
	flagNoCopy    bool
)

func init() {
	flag.Usage = func() {
		fmt.Print("\nUtility to find the best scoring compounds\n")
		fmt.Print("from a virtual screen with smina.\n\n")
		fmt.Print("Written in Go, COMAS 2018, license: MIT\n\n")
		fmt.Printf("Usage:  %s [options] <dir with scoring results from vs> <dir to write results>\n\n", os.Args[0])
		fmt.Println("Available options:")
		flag.PrintDefaults()
		fmt.Println("\nOne of the options --highest or --score has to be used!")
	}
	flag.IntVar(&flagTopN, "top", 0, "get n ligands with lowest values\n(outcome depends on sorting property, see --sortby)")
	flag.IntVar(&flagTopNperHA, "topha", 0, "get n ligands with lowest values per heavy atom\n(outcome depends on sorting property, see --sortby)")
	flag.Float64Var(&flagMaxValue, "maxval", 0.0, "get ligands with valaues lower than n\n(outcome depends on sorting property, see --sortby)")
	flag.IntVar(&flagNumHAmin, "minha", 0, "limit to ligands with min n heavy atoms\n(allows exclusion of fragment-like hits)")
	flag.IntVar(&flagNumHAmax, "maxha", numHAmaxDefault, "limit to ligands with max n heavy atoms")
	flag.StringVar(&flagSortBy, "sortby", "score", "property by which to sort the ligands\n(score (default) or le (Ligand Efficiency))")
	flag.BoolVar(&flagNoCopy, "nocopy", false, "do NOT copy the reported ligands to the result dir")
	flag.BoolVar(&flagAll, "all", false, "generate a score table for all ligands in the input dir; implies --nocopy.")
}

func getDir(flagPos int, msg string) string {
	result := flag.Arg(flagPos)
	if len(result) == 0 {
		fmt.Printf("\nMissing required positional argument: %s.\n", msg)
		flag.Usage()
		os.Exit(1)
	}
	return result
}

// getNumHeavyAtoms reads the pdbqt file and counts the number of heavy atoms
func getNumHeavyAtoms(scoreDir, logFile string) uint8 {
	var result uint8
	pdbqtFile := strings.Replace(logFile, ".log", ".pdbqt", -1)
	f, err := os.Open(filepath.Join(scoreDir, pdbqtFile))
	if err != nil {
		return 0
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ENDMDL") {
			return result
		}
		if strings.HasPrefix(line, "ATOM") {
			fields := strings.Fields(line)
			if fields[2] != "H" {
				result++
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return 0
	}
	return result
}

func getScore(scoreDir, logFile string) Score {
	ligID := strings.TrimSuffix(logFile, ".log")
	logContent, err := ioutil.ReadFile(filepath.Join(scoreDir, logFile))
	if err != nil {
		return Score{ligID, 0.0, 0, statusReadErr}
	}
	logText := string(logContent)
	start := strings.Index(logText, "1    ")
	if start < 0 {
		return Score{ligID, 0.0, 0, statusStartErr}
	}
	stop := strings.Index(logText[start:], "\n")
	if stop < 0 {
		return Score{ligID, 0.0, 0, statusStopErr}
	}
	// fmt.Println(start, stop)
	fields := strings.Fields(logText[start : start+stop])
	if len(fields) < 2 {
		return Score{ligID, 0.0, 0, statusFieldsErr}
	}
	// fmt.Printf("Fields are: %q", fields)
	val64, err := strconv.ParseFloat(fields[1], 32)
	utls.QuitOnError(err)
	value := float32(val64)
	numHA := getNumHeavyAtoms(scoreDir, logFile)
	if numHA == 0 {
		return Score{ligID, 0.0, 0, statusNumHAErr}
	}
	if LessF32(value, -20.0) || LessF32(-1.0, value) {
		return Score{ligID, value, numHA, statusImplausErr}
	}
	if LessF32(minScore, value) {
		return Score{ligID, value, numHA, statusLowScore}
	}
	return Score{ligID, value, numHA, statusOk}
}

// scanScores scans the Smina logfiles and returns:
// a list of ligand scores,
// the maximum number of heavy atoms in the ligands of the list,
// the minimum number of heavy atoms
func scanScores(scoreDir string) ([]Score, uint8, uint8) {
	fileCtr := 0
	files, err := ioutil.ReadDir(scoreDir)
	utls.QuitOnError(err)
	lock := &sync.Mutex{}
	result := []Score{}
	var maxHA uint8
	var minHA uint8 = 255
	statusCtrMap := make(map[uint8]int)
	limit := limiter.NewConcurrencyLimiter(numGoroutines)
	fmt.Println("")
	for _, file := range files {
		if logFile := file.Name(); strings.HasSuffix(logFile, ".log") {
			fileCtr++
			limit.Execute(func() {
				score := getScore(scoreDir, logFile)
				lock.Lock()
				statusCtrMap[score.status]++
				// Use all valid values when --all is set
				if flagAll {
					if score.status == statusOk || score.status == statusLowScore || score.status == statusImplausErr {
						result = append(result, score)
					}
					// Otherwise just use values with statusOk
				} else {
					if score.status == statusOk {
						result = append(result, score)
						maxHA = calculators.MaxUI8(maxHA, score.numHAtoms)
						minHA = calculators.MinUI8(minHA, score.numHAtoms)
					}
				}
				lock.Unlock()
			})
		}
	}
	limit.Wait()

	fmt.Printf("\nLogfiles scanned:   %6d\n", fileCtr)
	fmt.Printf("statusOk:           %6d\n", statusCtrMap[statusOk])
	fmt.Printf("statusLowScore:     %6d\n", statusCtrMap[statusLowScore])
	fmt.Printf("statusReadErr:      %6d\n", statusCtrMap[statusReadErr])
	fmt.Printf("statusStartErr:     %6d\n", statusCtrMap[statusStartErr])
	fmt.Printf("statusStopErr:      %6d\n", statusCtrMap[statusStopErr])
	fmt.Printf("statusFieldsErr:    %6d\n", statusCtrMap[statusFieldsErr])
	fmt.Printf("statusNumHAErr:     %6d\n", statusCtrMap[statusNumHAErr])
	fmt.Printf("statusImplausErr:   %6d\n", statusCtrMap[statusImplausErr])

	// if numScores+lowScores != fileCtr {
	// 	log.Println("Numbers do not add up!")
	// }
	return result, maxHA, minHA
}

// getValue returns the ligands value used for --sortBy
func getValue(s Score) float32 {
	if flagSortBy == "score" {
		return s.value
	}
	result, _ := calculators.LigEffUI8(s.value, s.numHAtoms)
	return result
}

func genReport(intro string, scores []Score) string {
	if len(scores) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(titleDefault + "\n" + intro + "\n")
	b.WriteString("Id\tScore\tNumHA\tLigEff\tRemark\n")
	for _, s := range scores {
		// Because the check that s.numHAtoms is gt 0 was performed already
		// in getNumHeavyAtoms(), the error can be ignored here
		ligEff, _ := calculators.LigEffUI8(s.value, s.numHAtoms)
		fmt.Fprintf(&b, "%s\t%.1f\t%d\t%.3f\n", s.id, s.value, s.numHAtoms, ligEff)
	}
	return b.String()
}

func copyLigands(scoreDir, resDir string, scores []Score) {
	for _, s := range scores {
		for _, ext := range []string{".pdbqt", ".log"} {
			src := filepath.Join(scoreDir, fmt.Sprintf("%s%s", s.id, ext))
			dst := filepath.Join(resDir, fmt.Sprintf("%s%s", s.id, ext))
			err := copy(src, dst)
			utls.QuitOnError(err)
		}
	}
}

func main() {
	flag.Parse()
	scoreDir := getDir(0, "directory with scoring results from vs")
	var resDir string
	if flagAll {
		flagNoCopy = true
	}
	if !flagNoCopy {
		resDir = getDir(1, "directory to write results")
		os.Mkdir(resDir, os.ModePerm)
	}
	var mode string
	if flagAll {
		mode = "all"
	} else {
		if flagTopNperHA > 0 {
			mode = "topha"
		} else if flagTopN != 0 {
			mode = "topn"
		} else if flagMaxValue < 0.0 {
			mode = "value"
		} else {
			fmt.Print("\nPlease use either --topha, --top or --maxval option.\n")
			flag.Usage()
			return
		}
	}
	// fmt.Println("Mode:", mode)
	// fmt.Println("NoCopy:", flagNoCopy)
	// fmt.Println("")

	scores, maxHA, minHA := scanScores(scoreDir)
	var (
		report       string
		sortProp     string
		subTitle     string
		head         []Score
		topNperHActr int
	)
	if flagSortBy == "le" {
		fmt.Println("\nSorting ligands by Ligand Efficiency...")
		sortProp = "ligand efficiency"
		sort.Sort(ByLE(scores))
	} else {
		sortProp = "score"
		sort.Sort(ByScore(scores))
	}

	limitNumHAmax8 := calculators.MinUI8(maxHA, uint8(flagNumHAmax))
	fmt.Println("maxHA, limitNumHAmax8:", maxHA, limitNumHAmax8)
	limitNumHAmin8 := calculators.MaxUI8(minHA, uint8(flagNumHAmin))
	if mode == "all" {
		report = genReport("Scores of all ligands", scores)
	} else {
		head = []Score{}
		if mode == "topha" {
			head = []Score{}
			for numHA := limitNumHAmax8; numHA >= limitNumHAmin8; numHA-- {
				ctr := 0
				for _, s := range scores {
					if s.numHAtoms == numHA {
						head = append(head, s)
						ctr++
						topNperHActr++
						if ctr >= flagTopNperHA {
							break
						}
					}
				}
			}
		} else {
			ctr := 0
			for _, s := range scores {
				if s.numHAtoms < limitNumHAmin8 {
					continue
				}
				if s.numHAtoms > limitNumHAmax8 {
					continue
				}
				if mode == "topn" {
					ctr++
					if ctr > flagTopN {
						break
					}
				} else {
					val := getValue(s)
					if val > float32(flagMaxValue) {
						break
					}
				}
				head = append(head, s)
			}
		}
		if mode == "topha" {
			subTitle = fmt.Sprintf("%d compounds with lowest %s values per heavy atom (lower is better; %d compounds in total)", flagTopNperHA, sortProp, topNperHActr)
		} else if mode == "topn" {
			subTitle = fmt.Sprintf("%d compounds with lowest %s values (lower is better)", flagTopN, sortProp)
		} else {
			subTitle = fmt.Sprintf("%d compounds with %s values <= %f (lower is better)", len(head), sortProp, flagMaxValue)
		}
		if mode == "topha" || flagNumHAmin > 0 {
			subTitle = subTitle + fmt.Sprintf("\nMinimum number of heavy atoms: %d", limitNumHAmin8)
		}
		if mode == "topha" || flagNumHAmax < numHAmaxDefault {
			subTitle = subTitle + fmt.Sprintf("\nMaximum number of heavy atoms: %d", limitNumHAmax8)
		}
		report = genReport(subTitle, head)
	}
	if !flagNoCopy {
		copyLigands(scoreDir, resDir, head)
	}
	fmt.Printf("\n\n")
	if !flagAll {
		fmt.Println(report)
	}
	ioutil.WriteFile("scores.txt", []byte(report), 0644)
}
