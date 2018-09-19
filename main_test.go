package main

import (
	"os"
	"testing"

	"github.com/tanelpuhu/go/str"
)

const mergeinfodata = `
------------------------------------------------------------------------
r12345 | username1 | 2000-01-01 00:00:53 +0000 (Wed, 12 Sep 2018) | 6 lines

JIRA-12345, Issue with script

Lorem Ipsum is simply dummy text of the printing and typesetting industry.
Lorem Ipsum has been the industry's standard dummy text ever since the 1500s,
when an unknown printer took a galley of type and scrambled it to make a type
specimen book. It has survived not only five centuries, but also the leap into
electronic typesetting....



------------------------------------------------------------------------
r12346 | user2 | 2002-01-01 00:00:58 +0000 (Wed, 19 Sep 2018) | 1 line

JIRA-12346 fixes
------------------------------------------------------------------------
r12347 | blaaaah | 2004-01-01 00:01:43 +0000 (Wed, 19 Sep 2018) | 1 line

JIRA-12347 more fixes
and some
and some
------------------------------------------------------------------------
`

func makeCommit() svnCommit {
	commit := svnCommit{}
	commit.revision = "1"
	commit.author = "me"
	commit.date = "2018-01-01"
	commit.msg = "fix for JIRA-334"
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

func TestParseMergeInfoLog(t *testing.T) {

	commits, err := parseMergeInfoLog([]byte(mergeinfodata))
	if err != nil {
		t.Fatalf("error parsing mergeinfo: %v", err)
	}
	if len(commits) != 3 {
		t.Errorf("did not find 3 but %d commits", len(commits))
	}
	expecting := []svnCommit{
		{"12345", "username1", "2000-01-01", "JIRA-12345, Issue with script"},
		{"12346", "user2", "2002-01-01", "JIRA-12346 fixes"},
		{"12347", "blaaaah", "2004-01-01", "JIRA-12347 more fixes"},
	}
	for i, commit := range commits {
		if commit.revision != expecting[i].revision {
			t.Errorf("unexpected revision, expected %s, got %s", expecting[i].revision, commit.revision)
		}
		if commit.author != expecting[i].author {
			t.Errorf("unexpected author, expected %s, got %s", expecting[i].author, commit.author)
		}
		if commit.date != expecting[i].date {
			t.Errorf("unexpected date, expected %s, got %s", expecting[i].date, commit.date)
		}
		if commit.msg != expecting[i].msg {
			t.Errorf("unexpected msg, expected %s, got %s", expecting[i].msg, commit.msg)
		}
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
