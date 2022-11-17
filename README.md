# Bumpr

<p align="center">
  <a href="https://github.com/PackagrIO/docs">
  <img width="300" alt="portfolio_view" src="https://github.com/PackagrIO/bumpr/raw/master/images/bumpr.png">
  </a>
</p>

Language agnostic tool to bump version files using SemVer. 

# Documentation
Full documentation is available at [PackagrIO/docs](https://github.com/PackagrIO/docs)


# Usage

```
cd /path/to/git/repo
cat pkg/version/version.go
# const VERSION = "0.0.3"

# export PACKAGR_PACKAGE_TYPE=[major/minor/patch]
packagr-bumpr start --scm github --package_type golang

cat pkg/version/version.go
# const VERSION = "0.0.4"
```

# Inputs
- `package_type`
- `scm`
- `version_bump_type`
- `version_metadata_path`
- `generic_version_template`
- `addl_version_metadata_paths`

# Outputs
- `release_version`

# Logo

- [chevron By Travis Avery, US ](https://thenounproject.com/travisavery/collection/ui-ux-circles-solid/?i=2453786)

