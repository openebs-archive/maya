package main

// GitCommit that was compiled. This will be filled in by the compiler.
var GitCommit string

// GitDescribe a commit using the most recent tag reachable from it.
var GitDescribe string

// Version show the version number,fill in by the compiler
var Version = "none"

// VersionPrerelease is a pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
const VersionPrerelease = "dev"
