#!/bin/bash

THIS_SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${THIS_SCRIPTDIR}/_utils.sh"
source "${THIS_SCRIPTDIR}/_formatted_output.sh"

# init / cleanup the formatted output
echo "" > "${formatted_output_file_path}"

write_section_start_to_formatted_output "# Updating CocoaPods"

write_section_start_to_formatted_output "## Current Cocoapods version"
pod_version=$(pod --version)
if [ $? -ne 0 ]; then
	write_section_to_formatted_output "# Error"
	write_section_start_to_formatted_output '* Failed to get current Cocoapods version'
	exit 1
fi

write_section_start_to_formatted_output "    ${pod_version}"

print_and_do_command_exit_on_error gem update cocoapods
print_and_do_command_exit_on_error pod setup --verbose

write_section_start_to_formatted_output "## Cocoapods version after update"
pod_version=$(pod --version)
if [ $? -ne 0 ]; then
	write_section_to_formatted_output "# Error"
	write_section_start_to_formatted_output '* Failed to get after-update Cocoapods version'
	exit 1
fi

write_section_start_to_formatted_output "    ${pod_version}"

exit 0
