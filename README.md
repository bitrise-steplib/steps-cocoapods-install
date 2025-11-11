# Run CocoaPods install

[![Step changelog](https://shields.io/github/v/release/bitrise-io/steps-cocoapods-install?include_prereleases&label=changelog&color=blueviolet)](https://github.com/bitrise-io/steps-cocoapods-install/releases)

This Step uses CocoaPods' `pod install` or `pod update` command to install your dependencies.

<details>
<summary>Description</summary>


CocoaPods is a dependency manager for Swift and Objective-C projects. This Step uses CocoaPods' `pod install` or `pod update` command to install your dependencies on the virtual machine where your Bitrise build runs.

CocoaPods version is determined based on the Podfile.lock file or on the Gemfile.lock file. If your Gemfile.lock file contains the `cocoapods` gem, then the Step will call the pod `install` command with `bundle exec`. Otherwise, the Cocoapods version in the Podfile.lock will be installed as a global gem.
If no Cocoapods version is defined in Podfile.lock or Gemfile.lock, the preinstalled sytem Cocoapods version will be used.

### Configuring the Step

1. Set the **Source Code Directory path** to the path of your app's source code.

1. Optionally, provide a Podfile in the **Podfile path** input.

   Without a specific Podfile, the Step does a recursive search for the Podfile in the root of your app's directory, and uses the first Podfile it finds.

### Troubleshooting

If the Step fails, check out the Podfile and the Gemfile of your app. Make sure there is no compatibility issue with the different versions of your Pods.

Check that both Podfile.lock and Gemfile.lock is committed and the Cocoapods versions defined in both match.

You can set the **Execute cocoapods in verbose mode?** input to true to get detailed logs of the Step.

### Useful links

* [Caching Cocoapods](https://devcenter.bitrise.io/builds/caching/caching-cocoapods/)
* [Include your dependencies in your repository](https://devcenter.bitrise.io/tips-and-tricks/optimize-your-build-times/#include-your-dependencies-in-your-repository)

### Related Steps

* [Run yarn command](https://www.bitrise.io/integrations/steps/yarn)
* [Carthage](https://www.bitrise.io/integrations/steps/carthage)
</details>

## 🧩 Get started

Add this step directly to your workflow in the [Bitrise Workflow Editor](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/steps/adding-steps-to-a-workflow.html).

You can also run this step directly with [Bitrise CLI](https://github.com/bitrise-io/bitrise).

## ⚙️ Configuration

<details>
<summary>Inputs</summary>

| Key | Description | Flags | Default |
| --- | --- | --- | --- |
| `command` | CocoaPods command to use for installing dependencies.  Available options: - `install`: Use `pod install` to download the explicit version listed in the Podfile.lock without trying to check if a newer version is available. - `update`: Use `pod update` to update every Pod listed in your Podfile to the latest version possible.  | required | `install` |
| `source_root_path` | Directory path where the project's Podfile (and optionally Gemfile) is placed.  CocoaPods commands will be executed in this directory.  | required | `$BITRISE_SOURCE_DIR` |
| `podfile_path` | Path of the project's Podfile.  By specifying this input `Workdir` gets overriden by the provided file's directory path. |  |  |
| `verbose` | Execute all CocoaPods commands in verbose mode.  If enabled the `--verbose` flag will be appended to all CocoaPods commands.  |  | `false` |
| `is_cache_disabled` | Disables automatic cache content collection.  By default the Step adds the Pods directory in the `Workdir` to the Bitrise Build Cache.  Set this input to disable automatic cache item collection for this Step.  |  | `false` |
</details>

<details>
<summary>Outputs</summary>
There are no outputs defined in this step
</details>

## 🙋 Contributing

We welcome [pull requests](https://github.com/bitrise-io/steps-cocoapods-install/pulls) and [issues](https://github.com/bitrise-io/steps-cocoapods-install/issues) against this repository.

For pull requests, work on your changes in a forked repository and use the Bitrise CLI to [run step tests locally](https://docs.bitrise.io/en/bitrise-ci/bitrise-cli/running-your-first-local-build-with-the-cli.html).

Learn more about developing steps:

- [Create your own step](https://docs.bitrise.io/en/bitrise-ci/workflows-and-pipelines/developing-your-own-bitrise-step/developing-a-new-step.html)
