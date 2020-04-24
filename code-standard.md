# Code Standards

## Sign your commits

We use the Developer Certificate of Origin (DCO) as an additional safeguard for the OpenEBS projects. This is a well established and widely used mechanism to assure that contributors have confirmed their right to license their contribution under the project's license. Please read [dcofile](https://github.com/openebs/openebs/blob/master/contribute/developer-certificate-of-origin). If you can certify it, then just add a line to every git commit message:

````
  Signed-off-by: Random J Developer <random@developer.example.org>
````

Use your real name (sorry, no pseudonyms or anonymous contributions). The email id should match the email id provided in your GitHub profile.
If you set your `user.name` and `user.email` in git config, you can sign your commit automatically with `git commit -s`.

## Verifying code style

We are using [golangci-lint](https://github.com/golangci/golangci-lint) to verify the linting errors. Make sure that your changes pass linting check by executing `make lint` or `make lint-docker` (docker version of `make lint`)

## Adding a changelog
If PR is about adding a new feature or bug fix then the Author of the PR is expected to add a changelog file with their pull request. This changelog file should be a new file created under the `changelogs/unreleased` folder. Name of this file must be in `pr_number-username` format and contents of the file should be the one-liner text which explains the feature or bug fix.

```sh
maya/changelogs/unreleased   <- folder
    12-github_user_name            <- file
```