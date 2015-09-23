package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var re = regexp.MustCompile("#(\\d+)")

func main() {
	argGhToken := flag.String("token", "", "(required) GitHub token to use in the API calls.")
	argUser := flag.String("user", "", "(required) Username owner of the repo.")
	argRepo := flag.String("repo", "", "(required) Repo to use.")
	argFromBranch := flag.String("from", "staging", "Source branch to compare from.")
	argToBranch := flag.String("to", "master", "Source branch to compare from.")

	flag.Parse()

	ghToken := *argGhToken

	if ghToken == "" {
		ghToken = os.Getenv("GH_TOKEN")
	}

	if ghToken == "" {
		fmt.Println("`token` not set.")
		os.Exit(1)
		return
	}
	if *argUser == "" {
		fmt.Println("`user` not set.")
		os.Exit(1)
		return
	}
	if *argRepo == "" {
		fmt.Println("`repo` not set.")
		os.Exit(1)
		return
	}

	desc, err := getDesc(ghToken, *argFromBranch, *argToBranch, *argUser, *argRepo)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
		return
	}
	fmt.Println(desc)
}

func getDesc(ghToken, from, to, user, repo string) (string, error) {
	response := &bytes.Buffer{}
	response.WriteString("Release\n")

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ghToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	a, _, err := client.Repositories.CompareCommits(user, repo, to, from)
	if err != nil {
		return "", err
	}

	issues, err := getIssues(ghToken, from, to, user, repo, client, a.Commits)
	if err != nil {
		return "", err
	}

	cc, err := getCc(a.Commits)
	if err != nil {
		return "", err
	}

	response.WriteString(issues)
	response.WriteString("\n")
	response.WriteString(cc)

	return response.String(), nil
}

func getIssues(ghToken, from, to, user, repo string, client *github.Client, commits []github.RepositoryCommit) (string, error) {
	response := &bytes.Buffer{}
	refIssues := make(map[int]bool)

	for _, commit := range commits {
		match := re.FindAllStringSubmatch(*commit.Commit.Message, -1)
		for _, issue := range match {
			id, _ := strconv.Atoi(issue[1])
			refIssues[id] = true
		}
	}

	var wg sync.WaitGroup
	var err error
	for k := range refIssues {
		wg.Add(1)
		go func(num int) {
			var issue *github.Issue
			issue, _, err = client.Issues.Get(user, repo, num)
			if err != nil {
				return
			}
			response.WriteString(fmt.Sprintf(
				"- [ ] #%d %v\n",
				num, *issue.Title))
			wg.Done()
		}(k)
	}
	wg.Wait()

	return response.String(), nil
}

func getCc(commits []github.RepositoryCommit) (string, error) {
	response := &bytes.Buffer{}
	users := make(map[string]bool)

	for _, commit := range commits {
		users[*commit.Committer.Login] = true
	}

	response.WriteString("/cc")
	for k := range users {
		response.WriteString(fmt.Sprintf(" @%v", k))
	}

	return response.String(), nil
}
