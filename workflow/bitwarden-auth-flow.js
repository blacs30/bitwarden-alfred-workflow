#!/usr/bin/osascript
ObjC.import('stdlib')
var BW_EXEC = $.getenv('BW_EXEC');
var PATH = $.getenv('PATH');

function run(arg) {

    var app = Application.currentApplication()
    app.includeStandardAdditions = true

    var mode = arg[0]
    var modeSet = "unlock"
    var email = arg[1]
    if (mode.localeCompare(modeSet) == 0) {
        var text = `Unlock Bitwarden for user ${email}.\nPlease enter your password:`
        var response = app.displayDialog(text, {
            defaultAnswer: "",
            withIcon: "caution",
            buttons: ["Cancel", "OK"],
            defaultButton: "OK",
            cancelButton: "Cancel",
            givingUpAfter: 120,
            hiddenAnswer: "true"
        })

        var password = response.textReturned
        return unlock(password)
    } else {
        var totpInt = arg[2]
        var totpStr = arg[3]
        if(totpStr == null){
            totpStr = " "
        }

        var text = `Login to Bitwarden for user ${email}.\nPlease enter your password:`
        var response = app.displayDialog(text, {
            defaultAnswer: "",
            withIcon: "caution",
            buttons: ["Cancel", "OK"],
            defaultButton: "OK",
            cancelButton: "Cancel",
            givingUpAfter: 120,
            hiddenAnswer: "true"
        })
        var password = response.textReturned

        if (totpStr.localeCompare(" ") != 0) {
            if (totpInt.localeCompare("1") != 0) {
                var text = `2FA authentication for Bitwarden user ${email}.\nPlease enter your 2FA code for ${totpStr}:`
                var response = app.displayDialog(text, {
                    defaultAnswer: "",
                    withIcon: "caution",
                    buttons: ["Cancel", "OK"],
                    defaultButton: "OK",
                    cancelButton: "Cancel",
                    givingUpAfter: 120,
                    hiddenAnswer: "false"
                })

                var totpCode = response.textReturned
                return login(email, password, totpCode, totpInt)
            } else {
                return login(email, password, "", totpInt)
            }
        } else {
            return login(email, password, "", "")
        }
    }
}

function login(email, password, totp, totpMode) {
    var escapedPassword = password.replace(/"/g, '\\"')

    var app = Application.currentApplication()
    app.includeStandardAdditions = true

    if (totpMode.localeCompare("1") == 0) {
        var cmd = `PATH=${PATH}; ${BW_EXEC} login ${email} --method ${totpMode} "${escapedPassword}"`
        try{
            var result = app.doShellScript(cmd);
        }catch(e) {
            var res = e.toString()
            res = res.includes("Two-step login code")
            if (!res) {
                console.log(e.toString())
                return e.toString()
            }
        }
        var text = `Email authentication for Bitwarden user ${email}.\nPlease enter your 2FA code sent via email:`
        var response = app.displayDialog(text, {
            defaultAnswer: "",
            withIcon: "caution",
            buttons: ["Cancel", "OK"],
            defaultButton: "OK",
            cancelButton: "Cancel",
            givingUpAfter: 120,
            hiddenAnswer: "false"
        })
        var totpCode = response.textReturned

        var cmd = `PATH=${PATH}; ${BW_EXEC} login ${email} --method ${totpMode} --code ${totpCode} "${escapedPassword}" --raw`
        try {
            var result = app.doShellScript(cmd);
        }catch(e) {
            console.log(e.toString())
            return e.toString()
        }
    } else {
        if (totp.localeCompare("") != 0) {
            var cmd = `PATH=${PATH}; ${BW_EXEC} login ${email} --method ${totpMode} --code ${totp} "${escapedPassword}" --raw`
        } else {
            var cmd = `PATH=${PATH}; ${BW_EXEC} login ${email} "${escapedPassword}" --raw`
        }
        try {
            var result = app.doShellScript(cmd);
        }catch(e) {
            console.log(e.toString())
            return e.toString()
        }
    }
    var res = setToken(result)
    if (res.localeCompare("") == 0) {
        return "Logged in."
    }
}

function unlock(password) {
    var escapedPassword = password.replace(/"/g, '\"')

    var app = Application.currentApplication()
    app.includeStandardAdditions = true
    var cmd = `PATH=${PATH}; ${BW_EXEC} unlock "${escapedPassword}" --raw`
    try {
        var result = app.doShellScript(cmd);
        var res = setToken(result)
        if (res.localeCompare("") == 0) {
            return "Unlocked"
        }
    }catch(e) {
        console.log(e)
        return e.toString()
    }
}

function setToken(token) {
    var bundleId = "com.lisowski-development.alfred.bitwarden"
    var app = Application.currentApplication()
    app.includeStandardAdditions = true

    var cmd = `/usr/bin/security add-generic-password -s ${bundleId} -a token -w ${token} -U`
    try {
        app.doShellScript(cmd);
        return ""
    }catch(e) {
        console.log(e.toString())
        return e.toString()
    }
}
