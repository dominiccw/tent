CHANGELOG
=========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.5] - 2018-11-27
## Fixed
- Fixed checking of nomad parsing and now checking for status codes less than 200 or greater than 299, instead of erroring on 200's.

## [1.0.4] - 2018-11-26
## Added
- Added more debug output when running in verbose mode to show the json document that is generated from the hcl nomad file.
## Changed
- Changed handling and detection of errors when parsing the nomad hcl file. There should now be less chance of a parse error going undetected.

## [1.0.3] - 2018-10-23
### Fixed
- Fixed issue where nomad deployments would fail but tent would exit with success.

## [1.0.2] - 2018-10-08
### Fixed
- Fixed issue where providing an invalid environment name when running deploy/destroy continued as normal instead or erroring.
- Fixed issue when attempting to deploy nomad batch jobs (these do not return an `EvalID`).

## 1.0.0 Initial Release
