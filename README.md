# Simple Bitwarden Workflow for Alfred

Simple yet powerful integration with the Bitwarden CLI so you can now get your passwords out of your Bitwarden vault and straight into the clipboard from within Alfred.

**Note**: Passwords with spaces at the beginning or end are _NOT_ supported

## Version 1.3.1 update - Please Read

-----
* Added ctrl+shift modifier to open the url of an item in the default browser

## Version 1.3.0 update - Please Read

-----
* Update workflow python package to version 1.37 with support for Alfred 4
* Add filtering support for Bitwarden folders #12

**Syntax:**

`bw -f folder_name search_string`

-- or --

`bw search_string -f folder_name`

-- or original syntax --

`bw search_string`

Thank you, @rustycamper, for your contribution!


## Version 1.2.4 update - Please Read

-----
* Uses utf-8 decoding now which fixes an issue where the json object could not be decoded and alfred bw would fail

## Version 1.2.3 update - Please Read

-----
* Fixes an issue where spaces within the item name causes the workflow to being unable to get the password/username/totp
* Removes newline at the end of the output

Thank you, @rasmusbe, for contributing.

## Version 1.2.2 update - Please Read

-----
Fixes an issue where spaces within the password prevent a user from login / unlock of the vault.

## Version 1.2.1 update - Please Read

-----
Fixes an issue where the login is successful but the workflow doesn't set the marker to save it but instead returns that the vault is locked.

## Version 1.2.0 update - Please Read

-----

Ladies and gents, I am happy to present v1.2.0 of the workflow.
As this workflow was originally a fork from the LastPass CLI it is now almost completely a rewritten codebase without using AppleScript calling an external applescript file to ask for the password. That is done now via inline AppleScript in Python.

All perl and main AppleScripts have been rewritten in Python.

If you haven't used Bitwarden before... you are crazy and you should! Say bye to LastPass and hello to selfhosting. It is the single greatest password manager package out there :D so check it out at [https://bitwarden.com](https://bitwarden.com).

## Version 1.1.0 update - Please Read

-----

Ladies and gents, I am happy to present v1.1.0 of the workflow. Before I continue, this workflow has not been developed from scratch. The LastPass CLI workflow was the start and was remodeled to fit the Bitwarden CLI. Nonetheless it was a SIGNIFICANT amount of work for me so if you like it and use it, please say thank you by donating towards my organic food. Any amount will do, whatever you feel the value is for you/your business/your time :)

I have never used LastPass, I prefer to selfhost my applications. From the day I heard about Bitwarden I loved it - that was at the beginning of this year (2018).

If you haven't used Bitwarden before... you are crazy and you should! Say bye to LastPass and hello to selfhosting. It is the single greatest password manager package out there :D so check it out at [https://bitwarden.com](https://bitwarden.com).

-----

## Donations
This workflow represents many many hours effort of development and testing. So if you love the workflow, and get use out of it every day, if you would like to donate as a thank you to buy me some healthy organic food (or organic coffee), or to put towards a shiny new gadget you can [donate to me via Paypal](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=K7BXYQ3SQ76J6).

<a href="https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=K7BXYQ3SQ76J6" target="_blank"><img src="https://www.paypalobjects.com/en_US/i/btn/btn_donate_SM.gif" border="0" alt="PayPal — The safer, easier way to pay online."></a>


## Installation

1. Ensure you have Alfred installed with the Alfred Powerpack License
2. Install Homebrew (if you do not have it already installed)
	1. You should be able to just run the command in a terminal window (as your own user account NOT with sudo)
	2. `ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`
	3. Alternatively visit http://brew.sh/ for further instructions.
3. Install Bitwarden CLI command line interface
	4. In a terminal window run
		`brew install bitwarden-cli`
5. Download the .alfredworkflow file
6. Open the .alfredworkflow file to import into Alfred
7. Run `bwsetemail yourloginemail@yourdomain.com` in Alfred to set your Bitwarden username.
8. Run `bwsetserver https://bitwarden.example.com` in Alfred to set your Bitwarden URL. Use https://bitwarden.com for the hosted bitwarden.

## Usage

* `bwsetemail yourname@example.com` - must be run when you first install/upgrade to version 1.0 or higher
* `bwsetemail` - Set the Bitwarden user account email
* `bwsetserver` - Set the Bitwarden server to connect to
* `bwset2fa` - Enable 2FA for Bitwarden login
* `bwset2famethod` - Set the method for the Bitwarden 2FA login (optional)
* `bwlogin` - Log in to Bitwarden
* `bwlogout` - Log out of Bitwarden
* `bwunlock` - Unlock the Bitwarden vault in case in case it is locked
* `bwsync` - Syncronize bitwarden with the remote server
* `bw <query>` Search Bitwarden vault for item containing <query>, press return to copy the password to clipboard.
* Shift modifier can be used on `bw <query>` to copy the username.
* Alt modifier can be used on `bw <query>` to copy the totp (if available).
* Ctrl+shift modifier can be used on `bw <query>` to open the url of an item (if available) in the default browser.

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D

## History

* Version 1.0.0 - Initial Release
* Version 1.0.1 - Fixed logout / not logged in warning
* Version 1.0.2 - Fixed erroring in case no username exist, catch the error correctly now.
* Version 1.1.0
* Version 1.2.0
* Version 1.2.1
* Version 1.2.2
* Version 1.2.3
* Version 1.3.1 - Added ctrl+shift modifier to open the url of an item in the default browser

## Credits

Created by [Claas Lisowski](https://lisowski-development.com). If you would like to get into contact you can do so via:
* [@blacs30 on Twitter](http://twitter.com/blacs30)
* [Claas Lisowski on LinkedIn](https://www.linkedin.com/in/claas-fridtjof-lisowski-558220b7/)

## License

Released under the GNU GENERAL PUBLIC LICENSE Version 2, June 1991

## Notes
NOTE: This Alfred Workflow is not affiliated in any way with Bitwarden. The Bitwarden trademark and logo are owned by Bitwarden.com. The Bitwarden logo and product name have been used with permission of the Bitwarden team.

My thanks go out to Bitwarden for their awesome product and the new CLI!
