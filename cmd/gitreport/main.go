package main

import (
	"os"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"time"

	"github.com/araddon/dateparse"
	gitpg "github.com/wojciechka/gitreport"
	"github.com/go-git/go-git/v5"
)

var (
	filenameRegexp = regexp.MustCompile("^.*?-LQ([A-Za-z0-9_-]+).txt$")

	queryFlag  = flag.String("query", "", "base64-encoded query or filename with query embedded to use")

	allFlag  	= flag.Bool("all", false, "query all refs and tags")
	sinceFlag  = flag.Duration("since", 0, "duration since when to query")
	fromFlag   = flag.String("from", "", "date and/or time since when to query")
	toFlag     = flag.String("to", "", "date and/or time up to when to query")
	authorFlag = flag.String("author", "", "regular expression to use for author matching")
	refFlag    = flag.String("ref", "", "reference, branch or head to use for getting logs")
	outputFlag = flag.String("output", ".", "directory to output report(s) in")
)

func main() {
	flag.Parse()

	var q *gitpg.LogQuery

	if len(*queryFlag) > 0 {
		// remove any paths
		queryString := filepath.Base(*queryFlag)

		result := filenameRegexp.FindStringSubmatch(queryString)
		if len(result) >= 2 {
			queryString = result[1]
		}

		var err error
		q, err = gitpg.ImportLogQueryFilename(queryString)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		q = &gitpg.LogQuery{}

		if *sinceFlag != 0 && len(*fromFlag) > 0 {
			log.Fatal("since and from cannot be specified at the same time")
		}

		if *sinceFlag != 0 {
			t := time.Now().Add(-(*sinceFlag))
			t = t.Truncate(time.Second)
			q.FromTime = &t
		}

		if len(*fromFlag) > 0 {
			t, err := dateparse.ParseAny(*fromFlag)
			if err != nil {
				log.Fatal(err)
			}

			t = t.Truncate(time.Second)
			q.FromTime = &t
		}

		if len(*toFlag) > 0 {
			t, err := dateparse.ParseAny(*toFlag)
			if err != nil {
				log.Fatal(err)
			}

			t = t.Truncate(time.Second)
			q.ToTime = &t
		} else {
			t := time.Now()
			t = t.Truncate(time.Second)
			q.ToTime = &t
		}

		if len(*authorFlag) > 0 {
			q.AuthorRegexp = *authorFlag
		}

		if *allFlag {
			q.All = true
		} else if len(*refFlag) > 0 {
			q.Ref = *refFlag
		}
	}

	repositories := flag.Args()
	if len(repositories) == 0 {
		log.Fatal("No paths to repositories were provided")
	}

	for _, repo := range repositories {
		repo, err := filepath.Abs(repo)
		if err != nil {
			log.Fatal(err)
		}
		repoName := filepath.Base(repo)

		// TODO: allow customizing
		if repoName == "" || repoName == "." {
			repoName = "repository"
		}

		fileName, err := gitpg.ExportLogQueryFilename(q)
		if err != nil {
			log.Fatal(err)
		}

		fileName = filepath.Join(*outputFlag, fmt.Sprintf("%s-LQ%s.txt", repoName, fileName))
		fmt.Println(fileName)

		r, err := git.PlainOpen(repo)
		if err != nil {
			log.Fatal(err)
		}

		commitLog, err := gitpg.QueryCommitLog(r, q)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Generating report for repository %s to file %s\n", repo, fileName)

		file, err := os.Create(fileName)
		if err != nil {
			log.Fatal(err)
		}

		err = gitpg.WriteCommitLogReport(commitLog, file)
		if err != nil {
			log.Fatal(err)
		}

		err = file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}
}
