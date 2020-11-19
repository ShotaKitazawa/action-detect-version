package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PullRequestsFiles []struct {
	Sha         string `json:"sha"`
	Filename    string `json:"filename"`
	Status      string `json:"status"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	Changes     int    `json:"changes"`
	BlobURL     string `json:"blob_url"`
	RawURL      string `json:"raw_url"`
	ContentsURL string `json:"contents_url"`
	Patch       string `json:"patch"`
}

func main() {
	// Get environment variables
	prUrl, err := getEnvOrErr("INPUT_PR_URL")
	exitWhenError(err)

	versionDir, err := getEnvOrErr("INPUT_DIR")
	exitWhenError(err)
	versionDirCleaned := filepath.Clean(versionDir)

	githubToken, err := getEnvOrErr("GITHUB_TOKEN")
	exitWhenError(err)

	// list PR files from GitHub API
	prfs, err := listPullRequestsFiles(prUrl, githubToken)
	exitWhenError(err)

	// get version from PR files
	version, err := getVersion(prfs, versionDirCleaned)
	exitWhenError(err)

	// output for GitHub Actions
	Output(version)
}

func exitWhenError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getEnvOrErr(str string) (string, error) {
	val := os.Getenv(str)
	if val == "" {
		return "", fmt.Errorf("%s: unbound variable", str)
	}
	return val, nil
}

func listPullRequestsFiles(url, token string) (PullRequestsFiles, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var prfs PullRequestsFiles
	err = json.Unmarshal(body, &prfs)
	if err != nil {
		return nil, err
	}

	return prfs, nil
}

func getVersion(prfs PullRequestsFiles, versionDir string) (string, error) {
	var version string
	for _, prf := range prfs {
		if strings.HasPrefix(prf.Filename, versionDir) {
			if version == "" {
				version = strings.Split(prf.Filename, "/")[0]
			} else {
				if version != strings.Split(prf.Filename, "/")[0] {
					return "", fmt.Errorf("error: updated multiple version")
				}
			}

		}
	}
	if version == "" {
		return "", fmt.Errorf("error: nothing updated")
	}
	return version, nil
}

func Output(version string) {
	fmt.Printf("::set-output name=new_version::%s\n", version)
}
