#
# This script "pre-loads" podspec repositories for Cocoapods
#  which are accessed through SSH.
#
# Related issue: https://github.com/CocoaPods/CocoaPods/issues/2534
#

require 'optparse'
require 'uri'


options = {
  podfile_path: nil
}
opt_parser = OptionParser.new do |opt|
  opt.banner = "Usage: cocoapods_ssh_source_fix.rb [OPTIONS]"
  opt.separator  ""
  opt.separator  "Options"

  opt.on("-f", "--podfile PODFILE_PATH", "Input podfile path") do |value|
    options[:podfile_path] = value
  end

  opt.on("-h","--help","help") do
    puts opt_parser
    exit
  end
end
opt_parser.parse!

unless options[:podfile_path]
  puts "[!] podfile_path is missing"
  puts opt_parser
  exit 1
end

puts "--- Config:"
puts options
puts "-----------"

# -------------------------
# --- MAIN

def pod_add_repo(repo_url_string, repo_alias_name)
  fix_cmd_str = "pod repo add #{repo_alias_name} #{repo_url_string}"

  # remove previously applied fix - if this fix script
  #  would be called multiple times
  system("rm -rf #{ENV['HOME']}/.cocoapods/repos/#{repo_alias_name}")

  # apply fix
  puts " (i) Apply fix command: $ #{fix_cmd_str}"
  unless system(fix_cmd_str)
    raise "Failed to add pod spec repo: #{repo_url_string}"
  end
end

$pod_repo_fix_counter = 0
def apply_source_fix(uri_str)
  puts " * [fix] applying fix for uri (#{uri_str})"
  $pod_repo_fix_counter += 1
  pod_add_repo(uri_str, "SourceFix-#{$pod_repo_fix_counter}")
end

puts
puts "-> Fixing sources in #{options[:podfile_path]}"
puts

podfile_abs_path = File.expand_path(options[:podfile_path])
puts " (i) podfile_abs_path: #{podfile_abs_path}"
File.open(podfile_abs_path, 'r').each_line do |line|
  line_strip = line.strip
  parts = line_strip.split(' ')
  if parts.size >= 2 && parts[0].downcase == 'source'
    puts "source: #{parts}"
    expected_uri_part = parts[1].gsub(/\'/, '').gsub(/\"/, '')
    puts "expected_uri_part: #{expected_uri_part}"
    begin
      uri = URI(expected_uri_part)
      uri_scheme = uri.scheme
      if uri_scheme == 'ssh'
        apply_source_fix(expected_uri_part)
      else
        puts " * [no-fix] uri (#{uri}) should be handled by pod install, no fix required."
      end
    rescue => ex
      puts " (i) URI could not be detected, applying fix. Exception was: #{ex}"
      apply_source_fix(expected_uri_part)
    end
  end
end

puts
puts "-> Finished with source fixes."
