#!/usr/bin/osascript
// https://developer.apple.com/library/archive/documentation/LanguagesUtilities/Conceptual/MacAutomationScriptingGuide/PromptforText.html#//apple_ref/doc/uid/TP40016239-CH80-SW1
function run(arg){

    var app = Application.currentApplication()
    app.includeStandardAdditions = true

    var text = `Bitwarden ${arg[0]} for user ${arg[1]}.\nPlease enter your ${arg[2]}:`
    var answer = `${arg[3]}`
    var response = app.displayDialog(text, {
        defaultAnswer: "",
        withIcon: "caution",
        buttons: ["Cancel", "OK"],
        defaultButton: "OK",
        cancelButton: "Cancel",
        givingUpAfter: 120,
        hiddenAnswer: answer
    })
    return response.textReturned
}
