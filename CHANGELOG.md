*In compliance with the [GPL-2.0](https://opensource.org/licenses/GPL-2.0) license: I declare that this version of the program contains my modifications, which can be seen through the usual "git" mechanism.*  


2022-09  
Contributor(s):  
Claas Lisowski  
>improve login unlock; open in webui (#137)* check if userid value is nil, set empty string* Move login and unlock into go code* Unlock and sync after api login; remove cache expire checks* add webui_url ui config* fix lint/vet errors  
>Unlock and sync after api login; remove cache expire checks  
>add webui_url config  
>Move login and unlock into go code  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2022-03  
Contributor(s):  
blacs30  
Claas Lisowski  
>add option to choose to copy or open url (#125)  
>add instruction to clear cache  
>add option to choose to copy or open url  
>Feature/support v1.21.1 skip types (#123)* skip listed types, support v1.21.1  
>skip listed types, support v1.21.1  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2022-02  
Contributor(s):  
Trevor Cohen  
>Update README to specify that path to node executable must be in PATH (#121)If your `bw` and `node` executables are in different directories (e.g.if you use Homebrew to install `bw` and `nvm` for Node) then both thosepaths need to be added to the PATH workflow variable.  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2021-12  
Contributor(s):  
blacs30  
Claas Lisowski  
>change default keys back to .bw prefix, don't run auto syncs only show hint  
>use both macos-11 and macos-11.15 for build envs  
>support apikey login (#116)* support apikey login* use both macos-11 and macos-11.15 for build envs* update workflows* update go.* files* update version to 2.4.1* Update Makefile* move go files to src directory* change default keys back to .bw prefix, don't run auto syncs only show hint* Update README.md  
>Update README.md  
>update readme  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2021-08  
Contributor(s):  
blacs30  
Claas Lisowski  
>Improve login, move assets (#101)* Update README keywords, remove login and unlock from go code* move assets icons scripts etc directly into the workflow dir  
>Update README keywords, remove login and unlock from go code  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2021-06  
Contributor(s):  
blacs30  
Claas Lisowski  
>remove separate cache cli option, merge it into sync. use vars for login and unlock  
>merge master into branch  
>remove separate cache cli option, merge it into sync. use vars for login and unlock (#79)  
>add autolock service, update autosync service  
>add background sync via a macos launchagent  
>Add background sync (#80)* add background sync via a macos launchagent  
>update readme and fix startup time check  
>add Auto lock service (#81)* remove separate cache cli option, merge it into sync. use vars for login and unlock* add background sync via a macos launchagent* update info.plist with vars for auto lock and auto sync* add autolock service, update autosync service* fix go lint* update readme and fix startup time check  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2021-02  
Contributor(s):  
Claas Lisowski  
>add 2 troubleshooting topics (#62)* add 2 troubleshooting topics* add steps to update encryption keys  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2020-12  
Contributor(s):  
Claas Lisowski  
>Fix/encryption and sync (#58)* Fix encryption process and disable sync per default.  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2020-10  
Contributor(s):  
Claas Lisowski  
Luke Hamburg  
>Fix link in README (#48)Fix link - remove extra 'w'  
>Add new feature to decryption secrets from file. (#52)* Add custom path for data.json and move the load function into main/config.go* update cli help text* Add new feature to decrypt secrets from data.json file.* Increment version to 2.2.0  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2020-09  
Contributor(s):  
Claas Lisowski  
>Add gif to demonstrate filter mode change. (#44)  
>Feature/custom actions and alfred filter (#43)* Remove dead code* Update readme: changes in filtering and modifier actions.* Remove previous search options.* Update workflow and its version to 2.1.0* Remove space in config.* add: title with user option; modifier actions; folder keyword* README update with thanks to contributors and a note regarding the original fork.* Make the linter happy  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2020-08  
Contributor(s):  
Claas Lisowski  
>Add Bitwarden Alfred workflow v2 changes. (#21)  
>Increment version to 2.0.4. Use envconfig now. (#34)  
>Update changes config names in readme. (#37)  
>smaller fixes, demo gif (#24)* Add demo gif to README.* Increase version to 2.0.1 and update workflow description* Add debug print when unexpected cmd error occurs. * Use correct message for locked error.* Use previous bundle id.  
>adding full path to osascript; updating version to 2.0.3 (#30)Add info that for Alfred 4 version 4.1 is requiredrollback to old js for the go internal js prompt - only cli not workflow related  
>Improve config (#39)* Add keywords to config type* Add describtion to readme on how to update the path + gif* Fix totp message; hardcode osascript path; fix issue for login without 2fa* Fix issue for login without 2fa* Update version to 2.0.5  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2020-07  
Contributor(s):  
Christian Niehaves  
blacs30  
>Added ctrl+shift modifier to open the url of an item in the default browser (#16)* Ctrl+Shift modifier to open the url* Ctrl+Shift modifier to open the url* Updated readme for ctrl+shift modifier* Updated readme for ctrl+shift modifier* Ctrl+shift modifier to open the url* Ctrl+shift modifier to open the url* Ctrl+shift modifier to open the url* Ctrl+shift modifier to open the url* Updated readme for ctrl+shift modifier* Cleared variables  
>fix for alfred v4.1, get the correct index of list  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2020-02  
Contributor(s):  
Sting  
>added inline code (#14)just makes it much easier to read... :)  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2020-01  
Contributor(s):  
blacs30  
>Release 1.2.5  
>Update README and workflow source + package.  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2019-03  
Contributor(s):  
blacs30  
>Update readme for version 1.2.4  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2019-01  
Contributor(s):  
Dominik Moritz  
Claas Lisowski  
>Update README.md  
>Merge pull request #6 from domoritz/patch-1Update README.md  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2018-09  
Contributor(s):  
blacs30  
>Update to version 1.2.3 - solution for issue #3  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2018-08  
Contributor(s):  
blacs30  
Claas Lisowski  
>Update to version 1.2.2. Fix issue where spaces in the password prevent from login.  
>Add 2FA option to bitwarden alfred workflow. Fix issue where the vault was locked but the workflow only reported none object found. Open Alfred search again after config has been set.  
>Update to version 1.2.0. Rewrite AppleScript and perl scripts to Python.  
>Update README.mdadd history for version 1.2.1  
>Update to version 1.2.1.Fix issue with the login flow.Remove old workflow folder.  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 


2018-07  
Contributor(s):  
blacs30  
>Update README.md wich the 1.0.2 version description.  
>Update README.mdAdded history version 1.0.1  
>Update Readme. bwsetserver is required to set the remote server. Add missing commands to the usage paragraph.  
>Initial checkin for version 1.0.0 of bitwarden cli alfred workflow.  
- - - - - - - - - - - - - - - - - - - - - - - - - - - 

