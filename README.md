# PullRequest generator

Generates the description for pull requests between two deployable branches (like `staging` and `production`)

This is useful when opening pull requests to ship new code from a staging branch to a production branch.

It takes the commits between the branches and shows:
- The issues that are merged
- CC all github users that have a commit

## Usage

`pull-request-generator -user acroca -repo pull-request-generator -from staging -to production -token MY_GH_TOKEN`
