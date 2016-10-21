package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	gitql "github.com/mvader/gitql/git"
	"github.com/mvader/gitql/sql"

	"gopkg.in/src-d/go-git.v4"
)

type CmdQuery struct {
	cmd

	Path string `short:"p" long:"path" description:"Path where the git repository is located"`
	Args struct {
		SQL string `positional-arg-name:"sql" required:"true" description:"SQL query to execute"`
	} `positional-args:"yes"`

	r  *git.Repository
	db sql.Database
}

func (c *CmdQuery) Execute(args []string) error {
	if err := c.validate(); err != nil {
		return err
	}

	if err := c.buildDatabase(); err != nil {
		return err
	}

	if err := c.executeQuery(); err != nil {
		return err
	}

	return nil
}

func (c *CmdQuery) validate() error {
	var err error
	c.Path, err = findDotGitFolder(c.Path)
	if err != nil {
		return err
	}

	return nil
}
func (c *CmdQuery) buildDatabase() error {
	c.print("opening %q repository...\n", c.Path)

	var err error
	c.r, err = git.NewFilesystemRepository(c.Path)
	if err != nil {
		return err
	}

	empty, err := c.r.IsEmpty()
	if err != nil {
		return err
	}

	if empty {
		return errors.New("error: the repository is empty")
	}

	head, err := c.r.Head()
	if err != nil {
		return err
	}

	c.print("current HEAD %q\n", head.Hash())

	name := filepath.Base(filepath.Join(c.Path, ".."))
	c.db = gitql.NewDatabase(name, c.r)
	return nil
}

func (c *CmdQuery) executeQuery() error {
	c.print("executing %q at %q\n", c.Args.SQL, c.db.Name())

	fmt.Println(c.Args.SQL)
	return nil
}

func findDotGitFolder(path string) (string, error) {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}

	git := filepath.Join(path, ".git")
	_, err := os.Stat(git)
	if err == nil {
		return git, nil
	}

	if !os.IsNotExist(err) {
		return "", err
	}

	next := filepath.Join(path, "..")
	if next == path {
		return "", errors.New("unable to find a git repository")
	}

	return findDotGitFolder(next)
}