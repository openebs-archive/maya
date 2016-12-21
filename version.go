package main

// The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string
var GitDescribe string

// The latest git tag will be filled in by the compiler
var Version string = "none"

// A pre-release marker for the version. If this is "" (empty string)
// then it means that it is a final release. Otherwise, this is a pre-release
// such as "dev" (in development), "beta", "rc1", etc.
const VersionPrerelease = "dev"
