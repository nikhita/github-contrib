# github-contrib

github-contrib is a tool to create a list of the following for a contributor across all repos in a Github organization.

1. Pull Requests created.
2. Issues created.
3. Pull Requests reviewed.

The output will be in the markdown format. You can copy and paste the output to a markdown file ([Sample Output](https://gist.github.com/nikhita/b31ab2bf33d00a5ce185b0850d61df57)) and proudly show it to others. :sunglasses:

## Installation

**Prerequisites**: Go version 1.7 or greater.

1. Get the code

```
$ go get github.com/nikhita/github-contrib
```

2. Build

```
$ cd $GOPATH/src/github.com/nikhita/github-contrib
$ go install
```

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
    	Mandatory GitHub API token.
  -v	print version and exit (shorthand).
  -version
    	print version and exit.
```

## License

github-contrib is licensed under the [MIT License](/LICENSE).
