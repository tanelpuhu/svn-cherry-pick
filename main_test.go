package main

import (
	"os"
	"testing"

	"github.com/tanelpuhu/go/str"
)

func makeCommit() svnCommit {
	commit := svnCommit{}
	commit.revision = "1"
	commit.author = "me"
	commit.date = "2018-01-01"
	commit.msg = "fix for JIRA-334"
	commit.source = "^/branches/hotfix-8"
	return commit

}

func TestCommitRevMatch(t *testing.T) {
	commit, args := makeCommit(), []string{"1", "2", "3"}
	if !commit.matchRevision(args) {
		t.Errorf("Should match: %s not in %s?", commit.revision, args)
	}
}

func TestCommitRevMisMatch(t *testing.T) {
	commit, args := makeCommit(), []string{"11", "12", "13"}
	if commit.matchRevision(args) {
		t.Errorf("Should not match: %s in %s", commit.revision, args)
	}
}

func TestCommitTicketMatch(t *testing.T) {
	commit, args := makeCommit(), []string{"JIRA-123", "JIRA-334"}
	if !commit.matchTicket(args) {
		t.Errorf("Should match: '%s' vs %s?", commit.msg, args)
	}
}

func TestCommitTicketMisMatch(t *testing.T) {
	commit, args := makeCommit(), []string{"JIRA-123"}
	if commit.matchTicket(args) {
		t.Errorf("Should not match: '%s' vs %s?", commit.msg, args)
	}
}

func TestParseArgs(t *testing.T) {
	var (
		filterCommit []string
		filterTicket []string
	)
	os.Args = append(os.Args, "trunk")
	_, filterCommit, filterTicket = parseArgs(
		[]string{"trunk", "12", "34", "6543324", "FIX-123", "blah-999"},
	)
	if len(filterCommit) != 3 {
		t.Errorf("filterCommit length should be 3, filterCommit is '%s'", filterCommit)
	}
	if len(filterTicket) != 2 {
		t.Errorf("filterTicket length should be 3, filterTicket is '%s'", filterTicket)
	}

	for _, key := range []string{"12", "34", "6543324"} {
		if !str.InSlice(filterCommit, key) {
			t.Errorf("Did not find revision %s in '%s'", key, filterCommit)
		}
		if str.InSlice(filterTicket, key) {
			t.Errorf("Did find revision %s in '%s'", key, filterTicket)
		}
	}

	for _, key := range []string{"FIX-123", "blah-999"} {
		if !str.InSlice(filterTicket, key) {
			t.Errorf("Did not find revision %s in '%s'", key, filterTicket)
		}

		if str.InSlice(filterCommit, key) {
			t.Errorf("Did find %s in '%s'", key, filterCommit)
		}
	}
}

func TestSortRevisions(t *testing.T) {
	var (
		filterCommit []string
	)
	os.Args = append(os.Args, "trunk")
	_, filterCommit, _ = parseArgs(
		[]string{"trunk", "917", "9", "10", "450", "99", "402", "999"},
	)
	if len(filterCommit) != 7 {
		t.Errorf("filterCommit length should be 7, filterCommit is '%s'", filterCommit)
	}
	for i, val := range []string{"9", "10", "99", "402", "450", "917", "999"} {
		if filterCommit[i] != val {
			t.Errorf("revision %d should be '%s' and not '%s'", i+1, val, filterCommit[i])
		}
	}
}
