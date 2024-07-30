package gitpg

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitLog struct {
	Commits []*object.Commit
}

func QueryCommitLog(repo *git.Repository, query *LogQuery) (*CommitLog, error) {
	var authorRegexp *regexp.Regexp

	result := CommitLog{}
	options := &git.LogOptions{
		Order: git.LogOrderCommitterTime,
	}

	if query.All {
		options.All = true
	} else {
		var ref *plumbing.Reference
		var err error

		if len(query.Ref) > 0 {
			ref, err = findFirstReference(repo, []string{
				query.Ref,
				"refs/remotes/" + query.Ref,
				"refs/heads/" + query.Ref,
				"refs/tags/" + query.Ref,
			})
		} else {
			ref, err = repo.Head()
		}
		if err != nil {
			return nil, err
		}

		options.From = ref.Hash()
	}

	if len(query.AuthorRegexp) > 0 {
		var err error

		authorRegexp, err = regexp.Compile(query.AuthorRegexp)
		if err != nil {
			return nil, err
		}
	}

	commitIterator, err := repo.Log(options)

	if err != nil {
		return nil, err
	}

	err = commitIterator.ForEach(func(c *object.Commit) error {
		if query.FromTime != nil && c.Author.When.Before(*query.FromTime) {
			return nil
		}

		if query.ToTime != nil && c.Author.When.After(*query.ToTime) {
			return nil
		}

		if authorRegexp != nil {
			if !authorRegexp.MatchString(c.Author.Email) {
				return nil
			}
		}

		result.Commits = append(result.Commits, c)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func WriteCommitLogReport(commitLog *CommitLog, out *os.File) error {
	// TODO: add versioning of outputs in the future

	for _, commit := range commitLog.Commits {
		commitReport, err := commitReport(commit)
		if err != nil {
			return err
		}

		out.WriteString(commitReport)
	}

	return nil
}

func CommitLogReport(commitLog *CommitLog) (string, error) {
	var sb strings.Builder

	// TODO: add versioning of outputs in the future

	for _, commit := range commitLog.Commits {
		commitReport, err := commitReport(commit)
		if err != nil {
			return "", err
		}

		sb.WriteString(commitReport)
	}

	return sb.String(), nil
}

func findFirstReference(repo *git.Repository, references []string) (*plumbing.Reference, error) {
	for _, reference := range references {
		ref, err := repo.Reference(plumbing.ReferenceName(reference), true)
		if err == nil {
			return ref, nil
		}
	}

	// return error for original reference if it has failed
	return repo.Reference(plumbing.ReferenceName(references[0]), true)
}

func commitReport(commit *object.Commit) (string, error) {
	parent, err := commit.Parent(0)
	if err != nil {
		// ignore the error and assume parent is nil at this point
		parent = nil
	}

	files, err := commitChangedFiles(commit, parent)
	if err != nil {
		return "", err
	}

	sort.Strings(files)

	when := commit.Author.When.Truncate(time.Second).UTC().Format(time.RFC3339)
	return fmt.Sprintf(
		"Commit %s\n"+
			"Date   %s\n"+
			"Author %s <%s>\n"+
			"\n"+
			"%s\n"+
			"\n"+
			"----\n"+
			"",
		commit.Hash.String(),
		when,
		commit.Author.Name, commit.Author.Email,
		strings.Join(files, "\n"),
	), nil
}

func commitToFileMap(commit *object.Commit, fileMap map[string]*object.File) error {
	fileIter, err := commit.Files()
	if err != nil {
		return err
	}

	err = fileIter.ForEach(func(f *object.File) error {
		fileMap[f.Name] = f
		return nil
	})

	return nil
}

func commitChangedFiles(commit *object.Commit, parent *object.Commit) ([]string, error) {
	result := []string{}
	fileDone := make(map[string]bool)
	oldFileMap := make(map[string]*object.File)
	newFileMap := make(map[string]*object.File)

	if parent != nil {
		err := commitToFileMap(parent, oldFileMap)
		if err != nil {
			return []string{}, err
		}
	}
	err := commitToFileMap(commit, newFileMap)
	if err != nil {
		return []string{}, err
	}

	formatFile := func(file *object.File) string {
		if file != nil {
			return fmt.Sprintf(
				"%s:%s",
				file.Mode.String(),
				file.Hash.String(),
			)
		} else {
			return "0000000:0000000000000000000000000000000000000000"
		}
	}

	comparer := func(name string) {
		if !fileDone[name] {
			oldFile := oldFileMap[name]
			newFile := newFileMap[name]
			filesEqual := false

			// compare if files on both sides are the same
			if oldFile != nil && newFile != nil {
				oldFileHash := [20]byte(oldFile.Hash)
				newFileHash := [20]byte(newFile.Hash)
				// TODO: improve comparison code
				if oldFile.Mode == newFile.Mode && oldFile.Size == newFile.Size && bytes.Equal(oldFileHash[:], newFileHash[:]) {
					filesEqual = true
				}
			}

			if !filesEqual {
				if oldFile != nil && newFile != nil {
					result = append(result, fmt.Sprintf(
						"%s change %s %s", name, formatFile(oldFile), formatFile(newFile),
					))
				} else if oldFile != nil {
					result = append(result, fmt.Sprintf(
						"%s delete %s", name, formatFile(oldFile),
					))
				} else if newFile != nil {
					result = append(result, fmt.Sprintf(
						"%s create %s", name, formatFile(newFile),
					))
				}
			}

			fileDone[name] = true
		}
	}

	for name, _ := range oldFileMap {
		comparer(name)
	}
	for name, _ := range newFileMap {
		comparer(name)
	}

	return result, nil
}
