#!/bin/bash

THIS_SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${THIS_SCRIPTDIR}/_bash_utils/utils.sh"
source "${THIS_SCRIPTDIR}/_bash_utils/formatted_output.sh"

# Validate parameters
echo "Configs:"
echo "  * source_root_path: $source_root_path"
echo "  * podfile_path: $podfile_path"
echo "  * [deprecated!] is_update_cocoapods: $is_update_cocoapods"
echo "  * install_cocoapods_version: $install_cocoapods_version"

if [ -z "${source_root_path}" ]; then
  write_section_to_formatted_output "# Error"
  write_section_start_to_formatted_output '* source_root_path input is missing'
  exit 1
fi
print_and_do_command_exit_on_error cd "${source_root_path}"

if [ -n "${install_cocoapods_version}" ] ; then
  if [[ "${install_cocoapods_version}" == "latest" ]] ; then
    # Install latest Cocoapods version

    print_and_do_command_exit_on_error bash "${THIS_SCRIPTDIR}/_steps-cocoapods-update/step.sh"
  else
    # Install desired Cocoapods version

    # "gem uninstall" is required if you want to use
    # an *older* cocoapods version than the preinstalled one
    # note: gem treats pre-release versions as older than
    # any release version!
    print_and_do_command_exit_on_error gem uninstall --all --executables cocoapods

    # install the version you want to use
    print_and_do_command_exit_on_error gem install cocoapods --version "${install_cocoapods_version}" --no-document
    print_and_do_command_exit_on_error pod setup --verbose
  fi
else
  # Deprecated - Install latest Cocoapods version
  if [[ "${is_update_cocoapods}" != "false" ]] ; then
    echo
    echo "[!] is_update_cocoapods is deprecated, use install_cocoapods_version input instead of this."
    echo

    print_and_do_command_exit_on_error bash "${THIS_SCRIPTDIR}/_steps-cocoapods-update/step.sh"
  fi
fi

write_section_to_formatted_output "# Run pod install"
print_and_do_command_exit_on_error bash "${THIS_SCRIPTDIR}/run_pod_install.sh"

exit 0
