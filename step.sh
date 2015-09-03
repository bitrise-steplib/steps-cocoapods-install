#!/bin/bash

THIS_SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${THIS_SCRIPTDIR}/_bash_utils/utils.sh"
source "${THIS_SCRIPTDIR}/_bash_utils/formatted_output.sh"

# init / cleanup the formatted output
echo "" > "${formatted_output_file_path}"


if [ -z "${source_root_path}" ]; then
  write_section_to_formatted_output "# Error"
  write_section_start_to_formatted_output '* source_root_path input is missing'
  exit 1
fi

# Update Cocoapods
if [[ "${is_update_cocoapods}" != "false" ]] ; then
  print_and_do_command_exit_on_error bash "${THIS_SCRIPTDIR}/_steps-cocoapods-update/step.sh"
else
  write_section_to_formatted_output "*Skipping Cocoapods version update*"
  pod_version=$(pod --version)
  if [ $? -ne 0 ]; then
  	write_section_to_formatted_output "# Error"
  	write_section_start_to_formatted_output '* Failed to get current Cocoapods version'
  	exit 1
  fi

  write_section_start_to_formatted_output "CocoaPods version: ${pod_version}"
fi

print_and_do_command_exit_on_error cd "${source_root_path}"

write_section_to_formatted_output "# Run pod install"
print_and_do_command_exit_on_error bash "${THIS_SCRIPTDIR}/run_pod_install.sh"

exit 0
