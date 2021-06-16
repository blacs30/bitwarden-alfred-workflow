#!/bin/bash

_end() {
	local r m
	r=$1
	m="$2"
	echo "result|$r|$m"
	exit $r
}

_generateCalIntervals() {
	local cnt=0
	local _timedef
	local re='^[0-9]+$'
	AUTOSYNC_TIMES=$(_get_var_from_plist "${infoplist}" variables.AUTOSYNC_TIMES | /usr/bin/tr ',' '\n')
	[ -n "$AUTOSYNC_TIMES" ] || _end 1 "AUTOSYNC_TIMES was not defined in workflow variables"
	#choose StartInterval or StartCalendarInterval mode
	if [[ $AUTOSYNC_TIMES =~ $re ]]; then
		SYNC_DESC="every ${AUTOSYNC_TIMES} seconds"
		/usr/bin/plutil -insert StartInterval -integer "${AUTOSYNC_TIMES}" "${plist_tmp_path}"
		(( cnt++ ))
	else
		SYNC_DESC="at "
		/usr/bin/plutil -replace StartCalendarInterval -xml '<array/>' "${plist_tmp_path}"
		while IFS=: read -r AUTO_HOUR AUTO_MIN; do
			_timedef="${AUTO_HOUR}:${AUTO_MIN}"
			if ! [[ $AUTO_HOUR =~ $re ]]; then
				echo "error: HOUR must be a number; skipping [${_timedef}]" 1>&2
				continue
			fi
			if ! [[ $AUTO_MIN =~ $re ]]; then
				echo "error: MIN must be a number; skipping [${_timedef}]" 1>&2
				continue
			fi
			if [ "$AUTO_HOUR" -lt 0 ] || [ "$AUTO_HOUR" -gt 23 ]; then
				echo "error: HOUR must be between 0-23; skipping [${_timedef}]"
				continue
			fi
			if [ "$AUTO_MIN" -lt 0 ] || [ "$AUTO_MIN" -gt 59 ]; then
				echo "error: MIN must be between 0-59; skipping [${_timedef}]"
				continue
			fi
			/usr/bin/plutil -insert StartCalendarInterval.${cnt} -xml '<dict/>' "${plist_tmp_path}"
			/usr/bin/plutil -insert StartCalendarInterval.${cnt}.Hour -integer "${AUTO_HOUR}" "${plist_tmp_path}"
			/usr/bin/plutil -insert StartCalendarInterval.${cnt}.Minute -integer "${AUTO_MIN}" "${plist_tmp_path}"
			(( cnt++ ))
			SYNC_DESC+="${_timedef} "
		done <<<"${AUTOSYNC_TIMES}"
	fi
	if [ "$cnt" -gt 0 ]; then
		echo "generated config with ${cnt} timers [${SYNC_DESC}]"
		return 0
	else
		return 1
	fi
}

_createPlist() {
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
	</array>
	<key>StandardErrorPath</key>
	<string>/tmp/$launchd_name.err</string>
	<key>StandardOutPath</key>
	<string>/tmp/$launchd_name.out</string>
</dict>
</plist>
EOF
}

_install_service() {
	_remove_service
	echo "generating new service configuration"
	_createPlist
	if ! _generateCalIntervals; then _end 1 "invalid timer configuration"; fi
	if ! /usr/bin/plutil -lint "${plist_tmp_path}"; then _end 1 "did not generate a valid plist"; fi
	if ! /bin/cp -f "${plist_tmp_path}" "${plist_path}"; then _end 1 "failed to copy LaunchAgent"; fi
	echo "loading service"
	/bin/launchctl bootstrap gui/$(id -u) "${plist_path}"
	if [ $? -eq 0 ]; then
		_end 0 "Service has been installed; sync will run ${SYNC_DESC}"
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

_install_symlink() {
	AUTOSYNC_SCRIPT_DIR=$(_get_var_from_plist "${infoplist}" variables.AUTOSYNC_SCRIPT_DIR)
	[ -n "${AUTOSYNC_SCRIPT_DIR}" ] || AUTOSYNC_SCRIPT_DIR=/usr/local/bin
	if /bin/ln -sfv "$wf_dir/${0##*/}" "${AUTOSYNC_SCRIPT_DIR}"; then
		_end 0 "Symlink created in ${AUTOSYNC_SCRIPT_DIR}"
	else
		_end 1 "Symlink could not be created in ${AUTOSYNC_SCRIPT_DIR}"
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
[ -e "${prefs}" ] || _end 1 "can't find Alfred prefs"
wf_basedir=$(_get_var_from_plist "${prefs}" current)/workflows
[ -e "${wf_basedir}" ] || _end 1 "can't find Alfred workflow dir"

alfred_app_bundleid=com.runningwithcrayons.Alfred
alfred_workflow_bundleid=com.lisowski-development.alfred.bitwarden
alfred_workflow_cache="$HOME/Library/Caches/${alfred_app_bundleid}/Workflow Data/${alfred_workflow_bundleid}"
alfred_workflow_data="$HOME/Library/Application Support/Alfred/Workflow Data/${alfred_workflow_bundleid}"
launchd_name=${alfred_workflow_bundleid}_autosync
plist_path="$HOME/Library/LaunchAgents/${launchd_name}.plist"
plist_tmp_path="$TMPDIR/${launchd_name}.plist"

#wf_basedir=$(_get_var_from_plist "$HOME/Library/Preferences/com.runningwithcrayons.Alfred-Preferences.plist" syncfolder)
infoplist=$(/usr/bin/find "${wf_basedir}" -name info.plist -depth 2 -exec /usr/bin/grep -H "<string>${alfred_workflow_bundleid}</string>" {} \; | /usr/bin/awk -F: '{ print $1 }')
[ -e "${infoplist}" ] || _end 1 "can't find Bitwarden v2 workflow"
wf_dir=${infoplist%/*}
if [ `uname -p` = "i386" ]; then 
    wf_bin="${wf_dir}/bitwarden-alfred-workflow-amd64"
else
    wf_bin="${wf_dir}/bitwarden-alfred-workflow-arm64"
fi
alfred_workflow_version=$(_get_var_from_plist "${infoplist}" version)
[ -n "${alfred_workflow_version}" ] || _end 1 "can't determine workflow version"
echo "found workflow v${alfred_workflow_version} at ${wf_dir}" 1>&2
bwpath=$(_get_var_from_plist "${infoplist}" variables.PATH)
[ -n "${bwpath}" ] || _end 1 "PATH variable not set in workflow"
bwexec=$(_get_var_from_plist "${infoplist}" variables.BW_EXEC)
bwauth_keyword=$(_get_var_from_plist "${infoplist}" variables.bwauth_keyword)

export alfred_workflow_bundleid
export alfred_workflow_cache
export alfred_workflow_data
export alfred_workflow_version
export bwauth_keyword
export PATH=${bwpath}
export BW_EXEC=${bwexec}

case $1 in
	(-i|--install)
		_install_service
		exit
		;;
	(-r|--remove)
		_remove_service
		[ $? -eq 0 ] && _end 0 "Autosync service has been removed"
		exit
		;;
	(-l|--link)
		_install_symlink
		exit
		;;
esac

if ! hash "${bwexec}" 2>/dev/null; then
	_end 1 "bw command not found, check PATH env variable"
fi

/usr/bin/xattr -d com.apple.quarantine "$wf_bin" 2>/dev/null
"$wf_bin" -sync -force
#"$wf_bin" -cache
#"$wf_bin" -icons
