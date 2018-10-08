CHANGELOG
=========

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2018-10-08
### Fixed
- Fixed issue where providing an invalid environment name when running deploy/destroy continued as normal instead or erroring.
- Fixed issue when attempting to deploy nomad batch jobs (these do not return an `EvalID`).

## 1.0.0 Initial Release
