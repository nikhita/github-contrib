# github-contrib

github-contrib is a tool to generate the following statistics of a contributor across _all_ public repositories in a Github organization.

1. Pull Requests created.
2. Issues created.
3. Pull Requests reviewed (but not created).

The output will be in the markdown format. You can copy and paste the output in a markdown file. [[Sample Output](https://gist.github.com/nikhita/b31ab2bf33d00a5ce185b0850d61df57)]

## Usage

Since Github enforces a rate limit on requests, you will need a personal API token. You can find more details about generating an API token [here](https://github.com/blog/1509-personal-api-tokens).

Please note that Github has a rate limit of 30 requests per minute for the search API so this will take some time fetch all results.
But don't worry, it should not take more than a few minutes! :smile:

```
$ github-contrib -h                                                                        
github-contrib : v0.1.0
USAGE:
github-contrib -token=<your-token> <org> <github-handle>
  -token string
    	GitHub API token. Mandatory.
  -v	print version and exit (shorthand).
  -version
    	print version and exit.
```

