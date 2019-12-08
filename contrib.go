package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
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
	flag.StringVar(&token, "token", "", "Mandatory GitHub API token")
	flag.BoolVar(&version, "version", false, "print version and exit")
	flag.BoolVar(&version, "v", false, "print version and exit (shorthand)")

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
	args := flag.Args()

	if len(args) != 2 {
		fmt.Println("Wrong number of arguments!")
		os.Exit(1)
	}

	org := args[0]
	author := args[1]

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

	dir := "./output/" + org
	file := author + ".md"
	f := createOutputFile(dir, file)
	defer f.Close()

	for _, repository := range repos {
		repo := repository.GetName()

		var output []string
		output = append(output, getCreatedPullRequests(ctx, client, org, repo, author)...)
		output = append(output, getIssues(ctx, client, org, repo, author)...)
		output = append(output, getReviewedPullRequests(ctx, client, org, repo, author)...)

		if len(output) != 0 {
			// for markdown-friendly output
			// TODO: refractor to be plain text friendly
			_, err := f.WriteString(fmt.Sprintf("**Repository: %s**\n", repo))
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			for _, line := range output {
				_, err := f.WriteString(line)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}

			f.WriteString("\n\n")
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
		createdPullRequests = append(createdPullRequests, fmt.Sprintf("\nTotal Pull Requests Created: %v", totalPullRequests))
	}

	for key, pr := range pullRequestResults.Issues {
		serialNumber := fmt.Sprintf("%v. ", key+1)
		pullRequestLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, pr.GetNumber(), pr.GetHTMLURL()) // org/repo#number
		pullRequestTitle := fmt.Sprintf("%s", pr.GetTitle())
		createdPullRequests = append(createdPullRequests, fmt.Sprintf("\n%s%s%s", serialNumber, pullRequestLink, pullRequestTitle))
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
		createdIssues = append(createdIssues, fmt.Sprintf("\nTotal Issues Opened: %v", totalIssues))
	}

	for key, issue := range issuesResults.Issues {
		serialNumber := fmt.Sprintf("%v. ", key+1)
		issueLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, issue.GetNumber(), issue.GetHTMLURL()) // org/repo#number
		issueTitle := fmt.Sprintf("%s", issue.GetTitle())
		createdIssues = append(createdIssues, fmt.Sprintf("\n%s%s%s", serialNumber, issueLink, issueTitle))
	}

	return createdIssues
}

// getReviewedPullRequests gets all Pull Requests reviewed by the author.
func getReviewedPullRequests(ctx context.Context, client *github.Client, org, repo, author string) []string {
	sleepIfRateLimitExceeded(ctx, client)
	var reviewedPullRequests []string

	allReviewedPullRequestsquery := "is:pr repo:" + org + "/" + repo + " reviewed-by:" + author + " -author:" + author
	opt := &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	reviewedPullRequestResults, _, err := client.Search.Issues(ctx, allReviewedPullRequestsquery, opt)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	totalReviewedPullRequests := reviewedPullRequestResults.GetTotal()
	if totalReviewedPullRequests != 0 {
		reviewedPullRequests = append(reviewedPullRequests, fmt.Sprintf("\nTotal Pull Requests Reviewed: %v", totalReviewedPullRequests))
	}

	for key, pr := range reviewedPullRequestResults.Issues {
		serialNumber := fmt.Sprintf("%v. ", key+1)
		pullRequestLink := fmt.Sprintf("[%s/%s#%v](%s) - ", org, repo, pr.GetNumber(), pr.GetHTMLURL()) // org/repo#number
		pullRequestTitle := fmt.Sprintf("%s", pr.GetTitle())
		reviewedPullRequests = append(reviewedPullRequests, fmt.Sprintf("\n%s%s%s", serialNumber, pullRequestLink, pullRequestTitle))
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
		timeToSleep := rateLimit.Search.Reset.Sub(time.Now()) + time.Second
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

func createOutputFile(dir, file string) (fp *os.File) {
	os.MkdirAll(dir, os.ModePerm)
	fp, err := os.Create(filepath.Join(dir, filepath.Base(file)))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return fp
}
