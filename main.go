package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin"
)

var (
	buildCommit string
)

func main() {
	fmt.Printf("Drone SVN Release Plugin built from %s\n", buildCommit)

	workspace := drone.Workspace{}
	repo := drone.Repo{}
	build := drone.Build{}
	vargs := Params{}

	plugin.Param("workspace", &workspace)
	plugin.Param("repo", &repo)
	plugin.Param("build", &build)
	plugin.Param("vargs", &vargs)
	plugin.MustParse()

	if build.Event != "tag" {
		fmt.Printf("The SVN Release plugin is only available for tags\n")
		os.Exit(0)
	}

	if workspace.Path != "" {
		os.Chdir(workspace.Path)
	}

	var files []string
	for _, glob := range vargs.Files.Slice() {
		globed, err := filepath.Glob(glob)
		if err != nil {
			fmt.Printf("Failed to glob %s\n", glob)
			os.Exit(1)
		}
		if globed != nil {
			files = append(files, globed...)
		}
	}

	if err := release(vargs.User, vargs.Password, vargs.BaseURL, files); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execute(cmd *exec.Cmd) error {
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	return cmd.Run()
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
