# lct: License Compliance Tool

Resolve all licenses used by your application, to verify your license compliance level. Currently supported platforms: Go.

## Install

```
$ go install github.com/ourstudio-se/lct@latest
```

## Basic usage

The basic usage prints out all licenses used by the dependencies specified in go.mod, transitively.

```
$ go mod graph | lct gomod 
```

## View dependency graph

A prettier graph visualization (including licenses) for dependencies, transitively.

```
$ go mod graph | lct gomod --graph
```

## View JSON

A parseable format which can be used by other tools.

```
$ go mod graph | lct gomod --json
```

## Bypass cache

```
$ go mod graph | lct gomod --no-cache
```

## Specify custom cache file

```
$ go mod graph | lct gomod --cache-file /path/to/file.cache
```

## Verification

### Verify a preset of allowed licenses

Create a rules definition file, `lct.rules.yaml`:
```
rules:
  allowed_licenses:
    - Apache-2.0
    - BSD-2-Clause
    - BSD-3-Clause
    - MIT
```
### Whitelisting package sources

Whitelist package sources such as non-public sources, or `replace` directives in `go.mod` by adding the `whitelisted_package_sources` directive:
```
rules:
  allowed_licenses:
    - Apache-2.0
    - BSD-2-Clause
    - BSD-3-Clause
    - MIT
  
  whitelisted_package_sources:
    - github.com/my-organization
    - github.com/some-repo/some-library
```
### Execute verification
Invoke `lct` with the `--verify-with` flag:

```
$ go mod graph | lct gomod --verify-with=lct.rules.yaml
```
