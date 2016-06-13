#!/bin/bash

THIS_SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${THIS_SCRIPTDIR}/_bash_utils/utils.sh"
source "${THIS_SCRIPTDIR}/_bash_utils/formatted_output.sh"

# Validate parameters
echo_string_to_formatted_output "Configs:"
echo_string_to_formatted_output "  * source_root_path: $source_root_path"
echo_string_to_formatted_output "  * podfile_path: $podfile_path"
echo_string_to_formatted_output "  * [deprecated!] is_update_cocoapods: $is_update_cocoapods"
echo_string_to_formatted_output "  * install_cocoapods_version: $install_cocoapods_version"

if [ -z "${source_root_path}" ]; then
  write_section_to_formatted_output "# Error"
  write_section_start_to_formatted_output '* source_root_path input is missing'
  exit 1
fi
print_and_do_command_exit_on_error cd "${source_root_path}"
echo

# Update cocoapods
update_version=""

if [ -n "${install_cocoapods_version}" ] ; then
  update_version="${install_cocoapods_version}"
elif [[ "${is_update_cocoapods}" != "false" ]] ; then
  # Deprecated - Install latest Cocoapods version
  echo_string_to_formatted_output "[!] is_update_cocoapods is deprecated, use install_cocoapods_version input instead of this."

  update_version="latest"
else
  version=""
  if [ -f "$source_root_path/Podfile.lock" ] ; then
    # COCOAPODS: 1.0.0
    regex="COCOAPODS: (.+)"
    while read line; do
      if [[ $line =~ $regex ]] ; then
        version="${BASH_REMATCH[1]}"
      fi
    done < "$source_root_path/Podfile.lock"
  fi

  if [ -n "$version" ] ; then
    update_version="$version"
  fi
fi

if [ -n "$update_version" ] ; then
  echo_string_to_formatted_output "update cocoapods to: $update_version"

  if [[ "${update_version}" == "latest" ]] ; then
     print_and_do_command_exit_on_error bash "${THIS_SCRIPTDIR}/_steps-cocoapods-update/step.sh"
  else
    # "gem uninstall" is required if you want to use
    # an *older* cocoapods version than the preinstalled one
    # note: gem treats pre-release versions as older than
    # any release version!
    print_and_do_command_exit_on_error gem uninstall --all --executables cocoapods

    # install the version you want to use
    print_and_do_command_exit_on_error gem install cocoapods --version "${install_cocoapods_version}" --no-document
    print_and_do_command_exit_on_error pod setup --verbose
  fi
fi

write_section_to_formatted_output "# Run pod install"
print_and_do_command_exit_on_error bash "${THIS_SCRIPTDIR}/run_pod_install.sh"

exit 0
