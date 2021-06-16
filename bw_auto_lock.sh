#!/bin/bash


_end() {
	local r m
	r=$1
	m="$2"
	echo "result|$r|$m"
	exit $r
}

_writeplist() {
/bin/cat <<EOF >"${plist_tmp_path}"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>$launchd_name</string>
	<key>ProgramArguments</key>
	<array>
		<string>/bin/bash</string>
		<string>--</string>
		<string>$wf_dir/${0##*/}</string>
		<string>$LOCK_TIMEOUT</string>
	</array>
	<key>RunAtLoad</key>
  <true/>
  <key>StartInterval</key>
  <integer>300</integer>
</dict>
</plist>
EOF
}

re='^[0-9]+$'
_install_service() {
  if ! [[ $LOCK_TIMEOUT =~ $re ]]; then
  	_end 1 "error: LOCK_TIMEOUT must be a number"
  fi

	_remove_service
	echo "generating new service configuration"
	_writeplist
	if ! /usr/bin/plutil -lint "${plist_tmp_path}"; then _end 1 "did not generate a valid plist"; fi
	if ! /bin/cp -f "${plist_tmp_path}" "${plist_path}"; then _end 1 "failed to copy LaunchAgent"; fi
	echo "loading service"
	/bin/launchctl bootstrap gui/$(id -u) "${plist_path}"
	if [ $? -eq 0 ]; then
		_end 0 "Service has been installed; autolock will lock after ${LOCK_TIMEOUT} minutes of inactivity"
	else
		_end 1 "Could not install service, check logfiles for detail"
	fi
}

_remove_service() {
	echo "removing existing service"
	/bin/launchctl bootout gui/$(id -u) "${plist_path}" 2>/dev/null
	/bin/rm "${plist_path}" 2>/dev/null
	if [ -e "${plist_path}" ]; then
		_end 1 "Could not delete existing LaunchAgent"
	else
		return 0
	fi
}

_get_var_from_plist() {
	# 1=filename, 2=key
	[ -n "$2" ] || return 1
	[ -e "$1" ] || return 1
	/usr/bin/plutil -extract "$2" xml1 -o - -- "$1" |
	/usr/bin/sed -n "s/.*<string>\(.*\)<\/string>.*/\1/p"
}

# find TMP dir
if [ -z "${TMPDIR}" ]; then
	TMPDIR=$(/usr/bin/getconf DARWIN_USER_TEMP_DIR)
	if [ ! -e "${TMPDIR}" ]; then
		_end 1 "could not find TMPDIR directory"
	fi
fi



prefs="$HOME/Library/Application Support/Alfred/prefs.json"
[ -e "${prefs}" ] || { echo "can't find Alfred prefs"; exit 1; }
wf_basedir=$(_get_var_from_plist "${prefs}" current)/workflows
[ -e "${wf_basedir}" ] || { echo "can't find Alfred workflow dir"; exit 1; }

alfred_app_bundleid=com.runningwithcrayons.Alfred
export alfred_workflow_bundleid=com.lisowski-development.alfred.bitwarden
export alfred_workflow_cache="$HOME/Library/Caches/${alfred_app_bundleid}/Workflow Data/${alfred_workflow_bundleid}"
export alfred_workflow_data="$HOME/Library/Application Support/Alfred/Workflow Data/${alfred_workflow_bundleid}"
launchd_name=${alfred_workflow_bundleid}_lock
plist_path="$HOME/Library/LaunchAgents/${0##*/}_lock_agent.plist"
plist_tmp_path="$TMPDIR/${launchd_name}.plist"


infoplist=$(/usr/bin/find "${wf_basedir}" -name info.plist -depth 2 -exec /usr/bin/grep -H "<string>${alfred_workflow_bundleid}</string>" {} \; | /usr/bin/awk -F: '{ print $1 }')
[ -e "${infoplist}" ] || { echo "can't find Bitwarden v2 workflow"; exit 1; }
wf_dir=${infoplist%/*}
alfred_workflow_version=$(_get_var_from_plist "${infoplist}" version)
[ -n "${alfred_workflow_version}" ] || { echo "can't determine workflow version"; exit 1; }
echo "found workflow v${alfred_workflow_version} at ${wf_dir}"
export alfred_workflow_version
bwpath=$(_get_var_from_plist "${infoplist}" variables.PATH)
[ -n "${bwpath}" ] || { echo "PATH variable not set in workflow"; exit 1; }
export PATH=${bwpath}
bwexec=$(_get_var_from_plist "${infoplist}" variables.BW_EXEC)
if ! hash "${bwexec}" 2>/dev/null; then
	echo "bw command not found, check PATH env variable"; exit 1;
fi
export BW_EXEC=${bwexec}

if [ `uname -p` = "i386" ]; then 
    wf_bin="${wf_dir}/bitwarden-alfred-workflow-amd64"
else
    wf_bin="${wf_dir}/bitwarden-alfred-workflow-arm64"
fi

case $1 in
	(-i|--install)
		_install_service
		exit
		;;
	(-r|--remove)
		_remove_service
		[ $? -eq 0 ] && _end 0 "Autolock service has been removed"
		exit
		;;
esac

# to lock bitwarden directly after start of the system we check the uptime
# if the system has just started within the last 10 minutes then lock bitwarden
uptime_string=$(/usr/sbin/sysctl -n kern.boottime | /usr/bin/awk '{print $4}' | /usr/bin/sed 's/,//')
now=$(/bin/date +%s)
if [ "$((now-uptime_string))" -lt 300 ]; then
  /usr/bin/xattr -d com.apple.quarantine "$wf_bin" 2>/dev/null
  "$wf_bin" -lock
fi

if [[ $1 =~ $re ]] && [[ -f "${alfred_workflow_cache}"/last-usage ]]; then
  last_usage=$(/bin/cat "${alfred_workflow_cache}"/last-usage)
	now=$(/bin/date +%s)
	if [ "$((now-last_usage))" -gt $(($1*60)) ]; then
    /usr/bin/xattr -d com.apple.quarantine "$wf_bin" 2>/dev/null
    "$wf_bin" -lock
  fi

fi
