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
	makeDir := exec.Command("make_svn_dir.sh", user, password, url, "svn-base-dir")
	if err := execute(makeDir); err != nil {
		return err
	}

	for _, file := range files {
		stage := exec.Command("cp", file, "svn-base-dir/")
		if err := execute(stage); err != nil {
			return fmt.Errorf("Failed to stage %s artifact: %s", file, err)
		}
	}

	push := exec.Command("push.sh", user, password, "svn-base-dir/")
	if err := execute(push); err != nil {
		return fmt.Errorf("Failed to stage artifacts: %s", err)
	}
	return nil
}
