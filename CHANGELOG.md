## Changelog (Current version: 1.4.2)

-----------------

### 1.4.2 (2016 Jun 14)

* [9ccc90d] prepare for version 1.4.2
* [c40cd5c] Merge pull request #9 from bitrise-io/install_fix
* [a182042] install version fix

### 1.4.1 (2016 Jun 13)

* [17112db] prepare for v1.4.1
* [028c30c] Merge pull request #8 from bitrise-io/cocoapods_version
* [c52598d] update cocoapods to Podfile.lock specified version

### 1.4.0 (2016 May 24)

* [757d200] release configs
* [032e6df] Merge pull request #7 from bitrise-io/install_cocoapods_version
* [6cca36c] install_cocoapods_versioninstall cocoapods version
* [62854f5] STEP_GIT_VERION_TAG_TO_SHARE: 1.3.0

### 1.3.0 (2016 May 21)

* [3bde379] `podfile_path` input handling
* [f369f47] bitrise.yml : better test setup - can be called separately and run all at the same time
* [43fe5b0] STEP_GIT_VERION_TAG_TO_SHARE: 1.2.1

### 1.2.1 (2016 May 19)

* [3cf12db] README update
* [fb17c83] step.yml cleanup
* [a4642a9] run `pod repo update` before retry - required for CocoaPods 1.0
* [5aa61bc] 1.2.0

### 1.2.0 (2016 Apr 13)

* [27ee14c] Try to `pod install --no-repo-update` first, and just retry without the `--no-repo-update` flag if that fails

### 1.1.0 (2016 Jan 11)

* [b2f6195] v1.1.0

### 1.0.3 (2015 Oct 05)

* [1a9b607] Merge pull request #3 from birmacher/master
* [c210005] don't search `Podfile` in `.git` directories
* [6153f86] Merge pull request #2 from bazscsa/patch-1
* [84caa91] depman-update
* [25e76cb] Update step.yml

### 1.0.2 (2015 Sep 21)

* [d2b3466] better testing `bitrise.yml`, with source-dir change (introduced in bitrise CLI 1.1.1) & README update
* [4241448] more (debug) information in case a Gemfile is found and used for `pod install` - printing the used `pod --version` as well, the one used through the Gemfile (`bundle exec`)
* [81a7d34] dont cleanup formatted output
* [9cbdd80] cd into the source root path as soon as possible

### 1.0.1 (2015 Sep 03)

* [4365217] renamed the depman managed cocoapods-update step's folder - prefixed with an underscore

-----------------

Updated: 2016 Jun 14