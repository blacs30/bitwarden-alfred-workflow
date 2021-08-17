#!/bin/bash

cd "$(dirname "$0")" || exit 1
/usr/bin/xattr -d com.apple.quarantine bitwarden-alfred-workflow 2>/dev/null
/bin/chmod +x bitwarden-alfred-workflow 2>/dev/null
