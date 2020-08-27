package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

var (
	// ErrNoSVN is given if svn is not in $PATH
	ErrNoSVN = errors.New("svn not present in $PATH")

	flagVersion bool
	flagUser    string
	flagSearch  string
)

const constVersion string = "0.0.8"
const constSVNMessageLimit = 120
const constSVNSepartatorLine = "------------------------------------------------------------------------"
const constSVNCommitLineRegex = `^r(\d*)\s\|\s([^\|]*)\s\|\s([^\|]*)\|\s(.*)$`
const constTicketRegex = `([A-Z]+-[0-9]+)`

type svnCommit struct {
	revision string
	author   string
	date     string
	msg      string
}

func stringInSlice(hey []string, needle string) bool {
	for _, item := range hey {
		if item == needle {
			return true
		}
	}
	return false
}

func (commit svnCommit) matchRevision(hay []string) bool {
	return stringInSlice(hay, commit.revision)
}

func (commit svnCommit) matchTicket(hay []string) bool {
	if len(hay) == 0 {
		return false
	}
	rex, _ := regexp.Compile(constTicketRegex)
	for _, item := range rex.FindStringSubmatch(commit.msg) {
		if stringInSlice(hay, item) {
			return true
		}
	}
	return false
}

func (commit svnCommit) CherryPick(source string) error {
	fmt.Printf("Cherrypicking r%s from %s...\n", commit.revision, source)
	cmd := exec.Command(
		"svn", "merge", "-c", commit.revision, source,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getEligibleCommits(source string) ([]svnCommit, error) {
	_, err := exec.LookPath("svn")
	if err != nil {
		return nil, ErrNoSVN
	}

	content, err := exec.Command(
		"svn", "mergeinfo", "--show-revs", "eligible", "--log", source, ".",
	).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", content)
	}
	return parseMergeInfoLog(content)
}

func parseMergeInfoLog(content []byte) ([]svnCommit, error) {
	var (
		res    = []svnCommit{}
		commit svnCommit
		line   string
	)
	rex, _ := regexp.Compile(constSVNCommitLineRegex)
	lines := strings.Split(string(content), "\n")
	for len(lines) != 0 {
		lines, line = lines[1:], lines[0]
		if line == constSVNSepartatorLine {
			lines, line = lines[1:], lines[0]
			resultSlice := rex.FindStringSubmatch(line)
			if len(resultSlice) == 5 {
				commit = svnCommit{}
				commit.revision = resultSlice[1]
				commit.author = resultSlice[2]
				commit.date = textToLocalTimeText(resultSlice[3][:25])
				// skip blank line, take first comment line
				lines, line = lines[2:], lines[1]
				commit.msg = line
				if len(commit.msg) > constSVNMessageLimit {
					commit.msg = commit.msg[:constSVNMessageLimit-3] + "..."
				}
				res = append(res, commit)
			}
		}
	}
	return res, nil
}

func textToLocalTimeText(text string) string {
	result, err := time.Parse("2006-01-02 15:04:05 -0700", text)
	if err != nil {
		log.Fatalf("error parsing date %s: %v", text, err)
	}
	return result.Local().Format("2006-01-02 15:04:05")
}

func parseArgs(args []string) (string, []string, []string) {
	var (
		filterCommit []string
		filterTicket []string
		arg          string
	)

	if len(args) > 1 {
		for i := 1; i < len(args); i++ {
			arg = args[i]
			_, err := strconv.Atoi(arg)
			if err == nil {
				filterCommit = append(filterCommit, arg)
			} else {
				filterTicket = append(filterTicket, arg)
			}
		}
	}
	sort.Slice(filterCommit, func(i, j int) bool {
		i, _ = strconv.Atoi(filterCommit[i])
		j, _ = strconv.Atoi(filterCommit[j])
		return i < j
	})
	source := flag.Arg(0)
	if !strings.HasPrefix(source, "^/") {
		if source != "trunk" && !strings.HasPrefix(source, "branches/") {
			source = "branches/" + source
		}
		source = "^/" + source
	}
	return source, filterCommit, filterTicket
}

func printUsage() {
	fmt.Fprint(os.Stderr, fmt.Sprintf(
		"Usage: %s [options] <source-path/branch-name/trunk> [revision-numbers and/or ticket-numbers]\n\n",
		filepath.Base(os.Args[0]),
	))
	fmt.Println("Options:")
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	flag.BoolVar(&flagVersion, "V", false, "Print version")
	flag.StringVar(&flagUser, "u", "", "filter by username")
	flag.StringVar(&flagSearch, "s", "", "filter by comment")
	flag.Usage = printUsage
	flag.Parse()
	if flagVersion {
		fmt.Printf("%s %v\n", filepath.Base(os.Args[0]), constVersion)
		os.Exit(0)
	}
	if flagSearch != "" {
		flagSearch = strings.ToLower(flagSearch)
	}

	var args []string
	args = flag.Args()
	if len(args) == 0 {
		flag.Usage()
	}
	source, filterCommit, filterTicket := parseArgs(args)
	commits, err := getEligibleCommits(source)
	if err != nil {
		log.Fatalf("Error getting eligible commits: %s\n", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	line := "%s\t%s\t%s\t%s"
	firstRow := false

	for _, commit := range commits {
		if len(filterTicket) > 0 && !commit.matchTicket(filterTicket) {
			continue
		}
		if flagUser != "" && commit.author != flagUser {
			continue
		}
		if flagSearch != "" && strings.Index(strings.ToLower(commit.msg), flagSearch) == -1 {
			continue
		}

		if len(filterCommit) > 0 {
			if commit.matchRevision(filterCommit) {
				err := commit.CherryPick(source)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			if firstRow == false {
				firstRow = true
				fmt.Fprintln(w, fmt.Sprintf(
					line,
					"Revision",
					"Author",
					"Date",
					"Message",
				))
			}
			fmt.Fprintln(w, fmt.Sprintf(
				line,
				commit.revision,
				commit.author,
				commit.date,
				commit.msg,
			))
		}
	}
	w.Flush()
}
