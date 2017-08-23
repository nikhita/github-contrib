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

const (
	// BANNER is what is printed for help/info output.
	BANNER = "github-contrib : %s\n"
	// USAGE is an example of how the command should be used.
	USAGE = "USAGE:\ngithub-contrib -token=<your-token> <org> <github-handle>"
	// VERSION is the binary version.
	VERSION = "v0.1.0"
)

var (
	token   string
	version bool
)

func init() {
	flag.StringVar(&token, "token", "", "GitHub API token. Mandatory.")
	flag.BoolVar(&version, "version", false, "print version and exit.")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand).")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, VERSION))
		fmt.Println(USAGE)
		flag.PrintDefaults()
	}

	flag.Parse()

	if version {
		fmt.Printf("%s", VERSION)
		os.Exit(0)
	}

	if token == "" {
		usageAndExit("GitHub token cannot be empty", 1)
	}
}

func main() {
	args := os.Args[1:]

	if len(args) != 3 {
		fmt.Println("Wrong number of arguments!")
		os.Exit(1)
	}

	org := args[1]
	author := args[2]

	ctx := context.Background()

	// Create an authenticated client.
	// Authenticated clients have a rate limit of 30 requests per minute.
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	getAllRepos(ctx, client, org, author)
}

// getAllRepos gets all stats for a contributor across all public repos in an org.
func getAllRepos(ctx context.Context, client *github.Client, org, author string) {
	opt := &github.RepositoryListByOrgOptions{Type: "public"}
	repos, _, err := client.Repositories.ListByOrg(ctx, org, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, repository := range repos {
		repo := repository.GetName()

		var output []string
		output = append(output, getCreatedPullRequests(ctx, client, org, repo, author)...)
		output = append(output, getIssues(ctx, client, org, repo, author)...)
		output = append(output, getReviewedPullRequests(ctx, client, org, repo, author)...)

		// for markdown-friendly output
		// TODO: refractor to be plain text friendly
		if len(output) != 0 {
			fmt.Printf("**Repository: %s**\n", repo)
			for _, line := range output {
				fmt.Println(line)
			}
			fmt.Printf("\n\n")
		}
	}
}

// getPullRequests gets all Pull Requests created by the author.
func getCreatedPullRequests(ctx context.Context, client *github.Client, org, repo, author string) []string {
	sleepIfRateLimitExceeded(ctx, client)
	var createdPullRequests []string

	allPullRequestsquery := "is:pr repo:" + org + "/" + repo + " author:" + author
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	pullRequestResults, _, err := client.Search.Issues(ctx, allPullRequestsquery, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	totalPullRequests := pullRequestResults.GetTotal()
	if totalPullRequests != 0 {
		createdPullRequests = append(createdPullRequests, fmt.Sprintf("\nTotal Pull Requests Created: %v", totalPullRequests))
	}

	for key, pr := range pullRequestResults.Issues {
		serialNumber := fmt.Sprintf("%v. ", key+1)
		pullRequestLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, pr.GetNumber(), pr.GetHTMLURL()) // org/repo#number
		pullRequestTitle := fmt.Sprintf("%s", pr.GetTitle())
		createdPullRequests = append(createdPullRequests, fmt.Sprintf("%s%s%s", serialNumber, pullRequestLink, pullRequestTitle))
	}

	return createdPullRequests
}

// getIssues gets all issues created by the author.
func getIssues(ctx context.Context, client *github.Client, org, repo, author string) []string {
	sleepIfRateLimitExceeded(ctx, client)
	var createdIssues []string

	allIssuesquery := "is:issue repo:" + org + "/" + repo + " author:" + author
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}

	issuesResults, _, err := client.Search.Issues(ctx, allIssuesquery, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	totalIssues := issuesResults.GetTotal()
	if totalIssues != 0 {
		createdIssues = append(createdIssues, fmt.Sprintf("\nTotal Issues Opened: %v", totalIssues))
	}

	for key, issue := range issuesResults.Issues {
		serialNumber := fmt.Sprintf("%v. ", key+1)
		issueLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, issue.GetNumber(), issue.GetHTMLURL()) // org/repo#number
		issueTitle := fmt.Sprintf("%s", issue.GetTitle())
		createdIssues = append(createdIssues, fmt.Sprintf("%s%s%s", serialNumber, issueLink, issueTitle))
	}

	return createdIssues
}

// getReviewedPullRequests gets all Pull Requests reviewed by the author.
// This does NOT include PRs created by the author.
func getReviewedPullRequests(ctx context.Context, client *github.Client, org, repo, author string) []string {
	sleepIfRateLimitExceeded(ctx, client)
	var reviewedPullRequests []string

	// this lists all pull requests reviewed (including the ones authored).
	allReviewedPullRequestsQuery := "is:pr repo:" + org + "/" + repo + " reviewed-by:" + author
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 50,
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
		reviewedPullRequests = append(reviewedPullRequests, fmt.Sprintf("\nTotal Pull Requests Reviewed: %v", totalReviewedAndNotAuthored))
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
			reviewedPullRequests = append(reviewedPullRequests, fmt.Sprintf("%s%s%s", serialNumber, pullRequestLink, pullRequestTitle))
			key++
		}
	}

	return reviewedPullRequests
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

func usageAndExit(message string, exitCode int) {
	if message != "" {
		fmt.Fprintf(os.Stderr, message)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(exitCode)
}
