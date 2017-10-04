package main

// GitCommit that was compiled.
// This will be filled in by the compiler.
var GitCommit string

// GitDescribe describe a commit.
var GitDescribe string

// Version has latest git tag which will be filled in by the compiler
var Version = "none"

// VersionPrerelease has A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
const VersionPrerelease = "dev"
