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

	"github.com/tanelpuhu/go/str"
)

// ErrNoSVN is given if svn is not in $PATH
var ErrNoSVN = errors.New("svn not present in $PATH")

const constVersion string = "0.0.5"
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

func (commit svnCommit) matchRevision(hay []string) bool {
	return str.InSlice(hay, commit.revision)
}

func (commit svnCommit) matchTicket(hay []string) bool {
	if len(hay) == 0 {
		return false
	}
	rex, _ := regexp.Compile(constTicketRegex)
	for _, item := range rex.FindStringSubmatch(commit.msg) {
		if str.InSlice(hay, item) {
			return true
		}
	}
	return false
}

func (commit svnCommit) CherryPick() error {
	fmt.Printf("Cherrypicking r%s from %s...\n", commit.revision, commit.source)
	content, err := exec.Command(
		"svn", "merge", "-c", commit.revision, commit.source,
	).CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%s", content)
	return nil
}

func getEligibleCommits(source string) ([]svnCommit, error) {
	var (
		res    = []svnCommit{}
		commit svnCommit
		line   string
	)

	_, err := exec.LookPath("svn")
	if err != nil {
		return res, ErrNoSVN
	}

	content, err := exec.Command(
		"svn", "mergeinfo", "--show-revs", "eligible", "--log", source, ".",
	).CombinedOutput()
	if err != nil {
		return res, err
	}
	rex, _ := regexp.Compile(constSVNCommitLineRegex)
	lines := strings.Split(string(content), "\n")
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
	return res, nil
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
	commits, err := getEligibleCommits(source)
	if err != nil {
		log.Fatalf("Error getting eligible commits: %s\n", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 2, 0, 2, ' ', 0)
	line := "%s\t%s\t%s\t%s"

	for _, commit := range commits {
		if len(filterTicket) > 0 && !commit.matchTicket(filterTicket) {
			continue
		}
		if len(filterCommit) > 0 {
			if commit.matchRevision(filterCommit) {
				err := commit.CherryPick()
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
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
