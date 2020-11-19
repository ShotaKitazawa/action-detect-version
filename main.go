package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
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
	prFilesUrl, err := pullRequestFilesURL(prUrl)
	exitWhenError(err)

	versionDir, err := getEnvOrErr("INPUT_DIR")
	exitWhenError(err)
	versionDirCleaned := filepath.Clean(versionDir) + "/"

	githubToken, err := getEnvOrErr("GITHUB_TOKEN")
	exitWhenError(err)

	// list PR files from GitHub API
	prfs, err := listPullRequestsFiles(prFilesUrl, githubToken)
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

func pullRequestFilesURL(prUrl string) (string, error) {
	u, err := url.Parse(prUrl)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, "files")
	return u.String(), nil
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http response code is not 200")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var prfs PullRequestsFiles
	if err = json.Unmarshal(body, &prfs); err != nil {
		return nil, err
	}

	return prfs, nil
}

func getVersion(prfs PullRequestsFiles, versionDir string) (string, error) {
	var version string
	for _, prf := range prfs {
		if strings.HasPrefix(prf.Filename, versionDir) {
			trimed := strings.TrimPrefix(prf.Filename, versionDir)
			if version == "" {
				version = strings.Split(trimed, "/")[0]
			} else {
				if version != strings.Split(trimed, "/")[0] {
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
