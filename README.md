# svn-cherry-pick

[![Build Status](https://travis-ci.org/tanelpuhu/svn-cherry-pick.svg?branch=master)](https://travis-ci.org/tanelpuhu/svn-cherry-pick)
[![Go Report Card](https://goreportcard.com/badge/github.com/tanelpuhu/svn-cherry-pick)](https://goreportcard.com/report/github.com/tanelpuhu/svn-cherry-pick)

Usage:

	svn-cherry-pick <branch-name> [revision-numbers] [ticket-numbers-(todo)]


Example:

	$ svn update
	Updating '.':
	At revision 9.
	$ svn revert -R .
	$ svn status
	$
	$ svn mergeinfo --show-revs eligible  ^/branches/branch-first
	r6
	r7
	r8
	$
	$ svn-cherry-pick branch-first
	    6        tanel 2018-03-24 lets show date
	    7        tanel 2018-03-24 func
	    8        tanel 2018-03-26 README
	$
	$ svn-cherry-pick branch-first 6 7
	Cherrypicking r6 from ^/branches/branch-first...
	--- Merging r6 into '.':
	U    app/scripts/demo.sh
	--- Recording mergeinfo for merge of r6 into '.':
	 U   .
	Cherrypicking r7 from ^/branches/branch-first...
	--- Merging r7 into '.':
	G    app/scripts/demo.sh
	--- Recording mergeinfo for merge of r7 into '.':
	 G   .
	$
	$ svn-cherry-pick branch-first
	    8        tanel 2018-03-26 README
	$
