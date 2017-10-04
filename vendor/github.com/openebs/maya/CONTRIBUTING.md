# Contributing to OpenEBS / Maya

Are you ready to enhance Maya. Great!! 

Feel free to contribute by:
- raising issues
- participate in commenting/voting on the issues
- create new PRs for Documentation or Code

## Contributing to Source Code


### Setting up your Development Environment

You will need an Linux host with Vagrant 1.9.1+ and Virtual Box 5.1. Follow these steps to setup your Development Environment.

```
cd <your-work-dir-on-linux-host>
git clone https://github.com/openebs/maya.git
vagrant up master-01
vagrant ssh master-01
#Only for the first time
make bootstrap
make deps
make dev
```


### Sign your work

We use the Developer Certificate of Origin (DCO) as a additional safeguard
for the OpenEBS project. This is a well established and widely used
mechanism to assure contributors have confirmed their right to license
their contribution under the project's license.
Please read [dcofile](https://github.com/openebs/openebs/blob/master/contribute/developer-certificate-of-origin).
If you can certify it, then just add a line to every git commit message:

````
  Signed-off-by: Random J Developer <random@developer.example.org>
````

Use your real name (sorry, no pseudonyms or anonymous contributions).
If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`. You can also use git [aliases](https://git-scm.com/book/tr/v2/Git-Basics-Git-Aliases)
like `git config --global alias.ci 'commit -s'`. Now you can commit with
`git ci` and the commit will be signed.
