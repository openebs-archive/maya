# Development Workflow

## Table of contents

- [Prerequisites](#prerequisites)
- [Initial Setup](#initial-setup)
- [Development](#development)
- [Dependencies](#dependencies)
- [Submitting Your Changes](#submitting-your-changes)

## Prerequisites

* You have Go 1.10+ installed on your local host/development machine.
* You have Docker installed on your local host/development machine. Docker is required for building maya component container images and to push them into a Kubernetes cluster for testing.

## Initial Setup

### Fork in the cloud

1. Visit https://github.com/openebs/maya.
2. Click `Fork` button (top right) to establish a cloud-based fork.

### Clone fork to local host

Place 'openebs/maya' code in any directory the following cloning procedure.
Create your clone:

```sh

mkdir path/to/directory/openebs
cd openebs

# Note: Here $user is your GitHub profile name
git clone https://github.com/$user/maya.git

# Configure remote upstream
cd openebs/maya
git remote add upstream https://github.com/openebs/maya.git

# Never push to upstream master
git remote set-url --push upstream no_push

# Confirm that your remotes make sense
git remote -v
```

## Development

### Always sync your local repository

Open a terminal on your local host. Change directory to the maya-fork root.

```sh

$ cd path/to/directory/openebs/maya
```

 Checkout the master branch.

 ```sh
 $ git checkout master
 Switched to branch 'master'
 Your branch is up-to-date with 'origin/master'.
 ```

 Recall that origin/master is a branch on your remote GitHub repository.
 Make sure you have the upstream remote openebs/maya by listing them.

 ```sh
 $ git remote -v
 origin	https://github.com/$user/maya.git (fetch)
 origin	https://github.com/$user/maya.git (push)
 upstream	https://github.com/openebs/maya.git (fetch)
 upstream	no_push (push)
 ```

 If the upstream is missing, add it by using below command.

 ```sh
 $ git remote add upstream https://github.com/openebs/maya.git
 ```

 Fetch all the changes from the upstream master branch.

 ```sh
 $ git fetch upstream master
 remote: Counting objects: 141, done.
 remote: Compressing objects: 100% (29/29), done.
 remote: Total 141 (delta 52), reused 46 (delta 46), pack-reused 66
 Receiving objects: 100% (141/141), 112.43 KiB | 0 bytes/s, done.
 Resolving deltas: 100% (79/79), done.
 From github.com:openebs/maya
   * branch            master     -> FETCH_HEAD
 ```

 Rebase your local master with the upstream/master.

 ```sh
 $ git rebase upstream/master
 First, rewinding head to replay your work on top of it...
 Fast-forwarded master to upstream/master.
 ```

 This command applies all the commits from the upstream master to your local master.

 Check the status of your local branch.

 ```sh
 $ git status
 On branch master
 Your branch is ahead of 'origin/master' by 38 commits.
 (use "git push" to publish your local commits)
 nothing to commit, working directory clean
 ```

 Your local repository now has all the changes from the upstream remote. You need to push the changes to your own remote fork which is origin master.

 Push the rebased master to origin master.

 ```sh
 $ git push origin master
 Username for 'https://github.com': $user
 Password for 'https://$user@github.com':
 Counting objects: 223, done.
 Compressing objects: 100% (38/38), done.
 Writing objects: 100% (69/69), 8.76 KiB | 0 bytes/s, done.
 Total 69 (delta 53), reused 47 (delta 31)
 To https://github.com/$user/maya.git
 8e107a9..5035fa1  master -> master
 ```

### Create a new feature branch to work on your issue

 Your branch name should have the format XX-descriptive where XX is the issue number you are working on followed by some descriptive text. For example:

 ```sh
 $ git checkout -b 1234-fix-developer-docs
 Switched to a new branch '1234-fix-developer-docs'
 ```

### Make your changes and build them

To build

 ```sh

 cd path/to/directory/openebs/maya
 GO111MODULE=on make mayactl
 ```

Check your linting.

 ```sh
 make golint
 ```

To build binaries for other distributions run the following command.

 ```sh
 make bin
 ```

### Test your changes

 ```sh
 cd $PATH/openebs/maya

 # Run every unit test
 make test
 ```

### Keep your branch in sync

[Rebasing](https://git-scm.com/docs/git-rebase) is very important to keep your branch in sync with the changes being made by others and to avoid huge merge conflicts while raising your Pull Requests. You will always have to rebase before raising the PR.

```sh
# While on your myfeature branch (see above)
git fetch upstream
git rebase upstream/master
```

While you rebase your changes, you must resolve any conflicts that might arise and build and test your changes using the above steps.

### Dependencies

#### Go modules

This repo supports [Go Modules](https://github.com/golang/go/wiki/Modules), and
is intended to be cloned outside the `GOPATH`, where Go Modules support is
enabled by default in Go 1.11.

If you are using this repo from within the `GOPATH`, activate module support
with:

```bash
export GO111MODULE=on
```

## Submitting your Changes

### Create a pull request

Before you raise the Pull Requests, ensure you have reviewed the checklist in the [CONTRIBUTING GUIDE](../CONTRIBUTING.md):

- Ensure that you have re-based your changes with the upstream using the steps above.
- Ensure that you have added the required unit tests for the bug fixes or new feature that you have introduced.
- Ensure your commits history is clean with proper header and descriptions.

Go to the [openebs/maya github](https://github.com/openebs/maya) and follow the Open Pull Request link to raise your PR from your development branch.
