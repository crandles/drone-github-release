package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type (
	Repo struct {
		Owner string
		Name  string
	}

	Build struct {
		Event string
	}

	Commit struct {
		Ref string
	}

	Config struct {
		User     string
		Files    []string
		Password string
		BaseURL  string
	}

	Plugin struct {
		Repo   Repo
		Build  Build
		Commit Commit
		Config Config
	}
)

func (p Plugin) Exec() error {
	var (
		files []string
	)

	// if p.Build.Event != "tag" {
	// 	return fmt.Errorf("The SVN Release plugin is only available for tags")
	// }

	if p.Config.User == "" {
		return fmt.Errorf("You must provide a User")
	}
	if p.Config.Password == "" {
		return fmt.Errorf("You must provide a Password")
	}
	if p.Config.BaseURL == "" {
		return fmt.Errorf("You must provide a SVN Directory URL")
	}

	for _, glob := range p.Config.Files {
		globed, err := filepath.Glob(glob)

		if err != nil {
			return fmt.Errorf("Failed to glob %s. %s", glob, err)
		}

		if globed != nil {
			files = append(files, globed...)
		}
	}

	if err := release(p.Config.User, p.Config.Password, p.Config.BaseURL, p.Config.Files); err != nil {
		return fmt.Errorf("Failed to upload the files. %s", err)
	}

	return nil
}

func release(user string, password string, url string, files []string) error {
	makeDir := exec.Command("svn", "co", "--no-auth-cache", "--username", user, "--password", password, "--depth", "empty", "--trust-server-cert", "--non-interactive", url, "svn-base-dir")
	if err := execute(makeDir); err != nil {
		return err
	}
	fmt.Println("directory created.")

	for _, file := range files {
		stage := exec.Command("cp", file, "svn-base-dir/")
		if err := execute(stage); err != nil {
			return fmt.Errorf("Failed to stage %s artifact: %s", file, err)
		}
	}

	if err := os.Chdir("svn-base-dir/"); err != nil {
		return fmt.Errorf("Failed to cd to svn directory: %s", err)
	}

	if err := execute(exec.Command("svn", "add", "*")); err != nil {
		return fmt.Errorf("Failed to add artifacts: %s", err)
	}

	if err := execute(exec.Command("svn", "ci", "--no-auth-cache", "--username", user, "--password", password, "--trust-server-cert", "--non-interactive", "-m", "$HOSTNAME: $DRONE_REPO-$DRONE_COMMIT: $DRONE_BUILD_NUMBER", "*")); err != nil {
		return fmt.Errorf("Failed to stage artifacts: %s", err)
	}

	return nil
}
