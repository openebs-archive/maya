# File Path Functions

While Sprig does not grant access to the filesystem, it does provide functions
for working with strings that follow file path conventions.

# base

Returns the last element of a path.

```
base "foo/bar/baz"
```

The above prints `baz`.

# dir

Returns the directory, stripping the last part of the path.

```
dir "foo/bar/baz"
```

The above returns `foo/bar`.

# clean

Cleans up the path.

```
clean "foo/bar/../baz"
```

The above resolves the `..` and returns `foo/baz`.

# ext

Returns the file extension.

```
ext "foo.bar"
```

The above returns `.bar`.

# isAbs

To check whether a file path is absolute or not, use `isAbs`.
