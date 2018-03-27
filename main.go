package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const constVersion string = "0.0.2"
const constSVNMessageLimit = 80
const constSVNSepartatorLine = "------------------------------------------------------------------------"
const constSVNCommitLineRegex = `^r(\d*)\s\|\s([^\|]*)\s\|\s([^\|]*)\|\s(.*)$`
const constTicketRegex = `([A-Z]+-[0-9]+)`

type svnCommit struct {
	revision string
	author   string
	date     string
	msg      string
	source   string
}

func (commit svnCommit) Print() {
	fmt.Printf("%5s %12s %s %s\n",
		commit.revision,
		commit.author,
		commit.date,
		commit.msg,
	)
}

func (commit svnCommit) matchRevision(hay []string) bool {
	return inStringSlice(hay, commit.revision)
}

func (commit svnCommit) matchTicket(hay []string) bool {
	if len(hay) == 0 {
		return false
	}
	rex, _ := regexp.Compile(constTicketRegex)
	for _, item := range rex.FindStringSubmatch(commit.msg) {
		if inStringSlice(hay, item) {
			return true
		}
	}
	return false
}

func errorAndExit(message string, error []byte) {
	fmt.Printf("Error getting mergeinfo:\n\n")
	lines := strings.Split(strings.Trim(string(error), "\n"), "\n")
	for _, line := range lines {
		fmt.Printf("=> %s\n", line)
	}
	fmt.Printf("\n")
	os.Exit(1)
}

func (commit svnCommit) CherryPick() {
	fmt.Printf("Cherrypicking r%s from %s...\n", commit.revision, commit.source)
	content, err := exec.Command(
		"svn", "merge", "-c", commit.revision, commit.source,
	).CombinedOutput()
	if err != nil {
		errorAndExit("Error cherry-picking", content)
	}
	fmt.Printf("%s", content)
}

func getEligibleCommits(source string) []svnCommit {
	var res = []svnCommit{}
	content, err := exec.Command(
		"svn", "mergeinfo", "--show-revs", "eligible", "--log", source, ".",
	).CombinedOutput()
	if err != nil {
		errorAndExit("Error getting mergeinfo", content)
	}
	rex, _ := regexp.Compile(constSVNCommitLineRegex)
	var commit svnCommit
	lines := strings.Split(string(content), "\n")
	var line string
	for len(lines) != 0 {
		lines, line = lines[1:], lines[0]
		if line == constSVNSepartatorLine {
			lines, line = lines[1:], lines[0]
			resultSlice := rex.FindStringSubmatch(line)
			if len(resultSlice) == 5 {
				commit = svnCommit{}
				commit.source = source
				commit.revision = resultSlice[1]
				commit.author = resultSlice[2]
				commit.date = resultSlice[3][:10]
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
	return res
}

func parseArgs(args []string) (string, []string, []string) {
	var filterCommit []string
	var filterTicket []string
	var arg string

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
	return flag.Arg(0), filterCommit, filterTicket

}

func inStringSlice(hay []string, needle string) bool {
	for _, item := range hay {
		if item == needle {
			return true
		}
	}
	return false
}

func init() {
	var flagVersion bool
	flag.BoolVar(&flagVersion, "V", false, "Print version")
	flag.Usage = func() {
		fmt.Fprint(os.Stderr,
			fmt.Sprintf("Usage: %s <source-path/branch-name/trunk>\n\n", filepath.Base(os.Args[0])))
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if flagVersion {
		fmt.Printf("%s %v\n", filepath.Base(os.Args[0]), constVersion)
		os.Exit(0)
	}
}

func main() {
	var args []string
	args = flag.Args()
	if len(args) == 0 {
		flag.Usage()
	}
	source, filterCommit, filterTicket := parseArgs(args)

	if !strings.HasPrefix(source, "^/") {
		if source != "trunk" && !strings.HasPrefix(source, "branches/") {
			source = "branches/" + source
		}
		source = "^/" + source
	}

	for _, commit := range getEligibleCommits(source) {
		if len(filterTicket) > 0 && !commit.matchTicket(filterTicket) {
			continue
		}
		if len(filterCommit) > 0 {
			if commit.matchRevision(filterCommit) {
				commit.CherryPick()
			}
		} else {
			commit.Print()
		}
	}
}
