package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

var token string

func init() {
	flag.StringVar(&token, "token", "", "GitHub API token")
	flag.Parse()
}

func main() {
	args := os.Args[1:]
	org := args[1]
	author := args[2]

	ctx := context.Background()

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	getAllRepos(ctx, client, org, author)
}

// getAllRepos gets all Pull Requests and Issues created and reviewed by the author.
func getAllRepos(ctx context.Context, client *github.Client, org, author string) {
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, _, err := client.Repositories.ListByOrg(ctx, org, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, repository := range repos {
		repo := repository.GetName()
		fmt.Printf("Repository: %s\n", repo)

		getCreatedPullRequests(ctx, client, org, repo, author)
		getIssues(ctx, client, org, repo, author)
		getReviewedPullRequests(ctx, client, org, repo, author)
	}
}

// getPullRequests gets all Pull Requests created by the author in the repo owned by the org.
func getCreatedPullRequests(ctx context.Context, client *github.Client, org, repo, author string) {
	sleepIfRateLimitExceeded(ctx, client)
	allPullRequestsquery := "is:pr repo:" + org + "/" + repo + " author:" + author
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	pullRequestResults, _, err := client.Search.Issues(ctx, allPullRequestsquery, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	totalPullRequests := pullRequestResults.GetTotal()
	if totalPullRequests != 0 {
		fmt.Println("Total Pull Requests Created: ", totalPullRequests)
	}

	for key, pr := range pullRequestResults.Issues {
		serialNumber := fmt.Sprintf("%v. ", key+1)
		pullRequestLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, pr.GetNumber(), pr.GetHTMLURL()) // org/repo#number
		pullRequestTitle := fmt.Sprintf("%s", pr.GetTitle())
		fmt.Println(serialNumber, pullRequestLink, pullRequestTitle)
	}
}

// getIssues gets all issues created by the author in the repo owned by the org.
func getIssues(ctx context.Context, client *github.Client, org, repo, author string) {
	sleepIfRateLimitExceeded(ctx, client)
	allIssuesquery := "is:issue repo:" + org + "/" + repo + " author:" + author
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	issuesResults, _, err := client.Search.Issues(ctx, allIssuesquery, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	totalIssues := issuesResults.GetTotal()
	if totalIssues != 0 {
		fmt.Println("Total Issues Opened: ", totalIssues)
	}

	for key, issue := range issuesResults.Issues {
		serialNumber := fmt.Sprintf("%v. ", key+1)
		issueLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, issue.GetNumber(), issue.GetHTMLURL()) // org/repo#number
		issueTitle := fmt.Sprintf("%s", issue.GetTitle())
		fmt.Println(serialNumber, issueLink, issueTitle)
	}
}

// getReviewedPullRequests gets all Pull Requests reviewed by the author in the repo owned by the org.
// This does not include PRs created by the author.
func getReviewedPullRequests(ctx context.Context, client *github.Client, org, repo, author string) {
	sleepIfRateLimitExceeded(ctx, client)
	// this lists all pull requests reviewed (including the ones authored).
	allReviewedPullRequestsQuery := "is:pr repo:" + org + "/" + repo + " reviewed-by:" + author
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}
	allReviewedResults, _, err := client.Search.Issues(ctx, allReviewedPullRequestsQuery, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	sleepIfRateLimitExceeded(ctx, client)
	reviewedAndAuthoredQuery := "is:pr repo:" + org + "/" + repo + " reviewed-by:" + author + " author:" + author
	reviewedAndAuthoredResults, _, err := client.Search.Issues(ctx, reviewedAndAuthoredQuery, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// this lists pull requests reviewed but NOT authored.
	totalReviewedAndNotAuthored := allReviewedResults.GetTotal() - reviewedAndAuthoredResults.GetTotal()
	if totalReviewedAndNotAuthored != 0 {
		fmt.Println("Total Pull Requests Reviewed: ", totalReviewedAndNotAuthored)
	}

	// mark authored Pull Requests as "seen".
	reviewedAndAuthored := make(map[int]bool, reviewedAndAuthoredResults.GetTotal())
	for _, authoredPR := range reviewedAndAuthoredResults.Issues {
		reviewedAndAuthored[authoredPR.GetNumber()] = true
	}

	key := 0
	for _, pr := range allReviewedResults.Issues {
		if !reviewedAndAuthored[pr.GetNumber()] {
			serialNumber := fmt.Sprintf("%v. ", key+1)
			pullRequestLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, pr.GetNumber(), pr.GetHTMLURL()) // org/repo#number
			pullRequestTitle := fmt.Sprintf("%s", pr.GetTitle())
			fmt.Println(serialNumber, pullRequestLink, pullRequestTitle)
			key++
		}
	}
}

func sleepIfRateLimitExceeded(ctx context.Context, client *github.Client) {
	rateLimit, _, err := client.RateLimits(ctx)
	if err != nil {
		fmt.Printf("Problem in getting rate limit information %v\n", err)
		return
	}

	if rateLimit.Search.Remaining == 1 {
		timeToSleep := rateLimit.Search.Reset.Sub(time.Now())
		time.Sleep(timeToSleep)
	}
}
