# Bitwarden Alfred Workflow

> Access your Bitwarden passwords, secrets, attachments and more via this powerful Alfred Workflow

## Features

* Completely rewritten in go
* fast secret / item search thanks to caching (no secrets are cached only the keys/names)
  * cache is encrypted
* access to (almost) all object information via this workflow
* download attachments via this workflow
* show favicons of the websites
* auto update
* uses the [awgo](https://pkg.go.dev/github.com/deanishe/awgo?tab=doc) framework/library
* many customizations possible


> Requires Alfred 4.1 or newer; NOT tested with Alfred 3

![Bitwarden V2 - Alfred Workflow Demo](./assets/bitwarden-v2.gif)

## Installation
- [Download the latest release](https://github.com/blacs30/bitwarden-alfred-workflow/releases)
- Open the downloaded file in Finder
- Make sure that the [Bitwarden CLI](https://github.com/bitwarden/cli#downloadinstall) is installed
- If running on macOS Catalina or later, you _**MUST**_ add Alfred to the list of security exceptions for running unsigned software. See [this guide](https://github.com/deanishe/awgo/wiki/Catalina) for instructions on how to do this.
  - <sub>Yes, this sucks and is annoying, but there is unfortunately is no easy way around this. macOS requires a paying Developer account for proper app notarization. I'm afraid I'm not willing to pay a yearly subscription fee to Apple just so that this (free and open source) project doesn't pester macOS Gatekeeper.</sub>

## PATH configuration

In many cases the bw executable is located in a path outside of the default system PATH.<br>
Please configure the Alfred Worklow variables PATH accordingly.<br>
In a normal terminal type `which bw` and copy the dirname (everything except the "bw") into the PATH workflow variable.<br>
The best is to append it to the existing string and separate it by a colon (:)

![Workflow PATH config](./assets/workflow-path-config.gif)

## Usage
To use, activate Alfred and type `.bw` to trigger this workflow. From there:

- type `.bwauth` for login/logout/unlock/lock
- type `.bwconfig` for settings/sync/workflow update/help/issue reports
- type any search term to search for secrets/notes/identities/cards
- modifier keys and actions are presented in the subtitle, different actions are available depending on the object type

## Search- / Filtermode

Up to version < 2.1.0 the *Fuzzy filtering a la Sublime Text* was default. Starting with version 2.1.0 Alfreds internal filtering is default.

You can change the search-/filtermode yourself easily. This gif shows the 3 steps which need to be done for it:
![Change filter mode](./assets/change-filter-mode.gif)

## Advanced Features / Configuration

- Configurable [workflow environment variables](https://www.alfredapp.com/help/workflows/advanced/variables/#environment)

| Name                      | Comment                                                                                                                                                                                               | Default Value                                                                       |
|---------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------|
| 2FA_ENABLED               | enables or disables 2FA for login (can be set via .bwconfig )                                                                                                                                         | true                                                                                |
| 2FA_NODE                  | sets the mode for the 2FA (can be set via .bwconfig ), 0 app, 1, email (not tested), 2 duo (not tested), 3 yubikey (not tested), 4 U2F (not tested)                                                   | 0                                                                                   |
| BW_EXEC                   | defines the binary/executable for the Bitwarden CLI command                                                                                                                                           | bw                                                                                  |
| bw_keyword                | defines the keyword which opens the Bitwarden Alfred Workflow                                                                                                                                         | .bw                                                                                 |
| bwf_keyword               | defines the keyword which opens the folder search of the Bitwarden Alfred Workflow                                                                                                                    | .bwf                                                                                |
| bwauth_keyword            | defines the keyword which opens the Bitwarden authentications of the Alfred Workflow                                                                                                                  | .bwauth                                                                             |
| bwconf_keyword            | defines the keyword which opens the Bitwarden configuration/settings of the Alfred Workflow                                                                                                           | .bwconfig                                                                           |
| EMAIL                     | the email which to use for the login via the Bitwarden CLI, will be read from the data.json of the Bitwarden CLI if present                                                                           | ""                                                                                  |
| EMPTY_DETAIL_RESULTS      | Show all information in the detail view, also if the content is empty                                                                                                                                 | false                                                                               |
| ICON_CACHE_ENABLED        | Download icons for login items if a URL is set                                                                                                                                                        | true                                                                                |
| ICON_CACHE_AGE            | This defines how old the icon cache can get in minutes, if expired the Workflow will download icons again. If icons are missing the workflow will also try to download them unrelated to this timeout | 43200 (1 month)                                                                     |
| AUTO_FETCH_ICON_CACHE_AGE | This defines how often the Workflow should check for an icon if is missing, it doesn't need to do it on every run hence this cache                                                                    | 1440 (1 day)                                                                        |
| MAX_RESULTS               | The number of items to display maximal in the search view                                                                                                                                             | 1000                                                                                |
| MODIFIER_1                | The first modifier key combination, possible options, which can be combined by comma separation, are "cmd,alt/opt,ctrl,shift,fn"                                                                      | alt                                                                                 |
| MODIFIER_2                | The first modifier key combination, possible options, which can be combined by comma separation, are "cmd,alt/opt,ctrl,shift,fn"                                                                      | shift                                                                               |
| MODIFIER_3                | The first modifier key combination, possible options, which can be combined by comma separation, are "cmd,alt/opt,ctrl,shift,fn"                                                                      | ctrl                                                                                |
| MODIFIER_4                | The first modifier key combination, possible options, which can be combined by comma separation, are "cmd,alt/opt,ctrl,shift,fn"                                                                      | cmd,opt                                                                             |
| MODIFIER_1_ACTION         | Action executed by the first modifier                                                                                                                                                                 | username,code                                                                       |
| MODIFIER_2_ACTION         | Action executed by the second modifier                                                                                                                                                                | url                                                                                 |
| MODIFIER_3_ACTION         | Action executed by the third modifier                                                                                                                                                                 | totp                                                                                |
| MODIFIER_4_ACTION         | Action executed by the fourth modifier                                                                                                                                                                | more                                                                                |
| NO_MODIFIER_ACTION        | Action executed without modifier pressed                                                                                                                                                              | password,card                                                                       |
| OUTPUT_FOLDER             | The folder to which attachments should be saved when the action is triggered. Default is \$HOME/Downloads. "~" can be used as well.                                                                   | ""                                                                                  |
| PATH                      | The PATH env variable which is used to search for executables (like the Bitwarden CLI configured with BW_EXEC, security to get and set keychain objects)                                              | /usr/bin:/usr/local/bin:/usr/local/sbin:/usr/local/share/npm/bin:/usr/bin:/usr/sbin |
| REORDERING_DISABLED       | If set to false the items which are often selected appear further up in the results.                                                                                                                  | true                                                                                |
| SERVER_URL                | Set the server url if you host your own Bitwarden instance - you can also set separate domains for api,webvault etc e.g. `--api http://localhost:4000 --identity http://localhost:33656`              | https://bitwarden.com                                                               |
| TITLE_WITH_USER           | If enabled the name of the login user item or the last 4 numbers of the card number will be appended (added) at the end of the name of the item                                                       | true                                                                        |
| SYNC_CACHE_AGE            | This defines how old the sync cache can get, if expired the Workflow will trigger a new sync with Bitwarden                                                                                           | 1440 (1 day)                                                                        |
| CACHE_AGE                 | This defines how old the cached items can get, if expired the Workflow will trigger a new cache refresh with Bitwarden (this does not involve a sync with the Bitwarden server)                       | 1440 (1 day)                                                                        |

## Modifier Actions Explained

| type     | action name                     |
|----------|---------------------------------|
| login    | password                        |
|          | username                        |
|          | url                             |
|          | totp                            |
| note     | - (always copy the secret note) |
| cards    | card                            |
|          | code                            |
| identity | - (always copy the name )       |
| others   | more (to show all item entries, can't be NO_MODIFIER_ACTION) |

You can place per type one `action name` into the ACTION config, a combination is possible where it is not overlapping with `more` or another of the same type.

Good examples:

NO_MODIFIER_ACTION=url,code
MODIFIER_1_ACTION=totp
MODIFIER_2_ACTION=more
MODIFIER_3_ACTION=password,card

Bad examples:

NO_MODIFIER_ACTION=url,password
MODIFIER_3_ACTION=code,card

# Develop locally

1. Install alfred cli <br>
`go get -u github.com/jason0x43/go-alfred/alfred`

2. Clone [this repo](https://github.com/blacs30/bitwarden-alfred-workflow).

3. Link the workflow directory with Alfred <br>
`cd workflow; alfred link`

4. Install dependency and run the first build<br>
`make build`

### Colors and Icons

*Light blue*

Hex: #175DDC <br>
RGB: 23,93,220

*Darker blue*

Hex: #134db7 <br>
RGB: 20,81,192

Get icons as pngs here https://fa2png.app/ and this is the browser https://fontawesome.com/cheatsheet


# Licensing and Thanks

The icons are based on [Bitwarden Brand](https://github.com/bitwarden/brand) , [Font Awesome](https://fontawesome.com/) and [Material Design](https://materialdesignicons.com/) Icons.

Parts of the README are taken over from [alfred-aws-console-services-workflow](https://github.com/rkoval/alfred-aws-console-services-workflow)

## Contributors

A big thanks to all code contributors but also to everyone who creates issues and helps that this workflow matures.

- @luckman212
- @blacs30

> Though this repository was a fork, it has 0 code reference anymore to the forked repo
> because of watchers and stars I decided to leave it this way and not to "unlink" it - by creating a new clean repository

## Source that helped me to get started

- [Writing Alfred workflows in Go](https://medium.com/@nikitavoloboev/writing-alfred-workflows-in-go-2a44f62dc432)
- [Example of the awgo package] (https://github.com/deanishe/awgo/blob/master/_examples/update/main.go)
- [awgo package](https://pkg.go.dev/github.com/deanishe/awgo?tab=doc)


## Troubleshooting

- "I'm seeing the following dialog when running the workflow"

  ![image](./assets/catalina-warning.png)

  Per [the installation steps](https://github.com/blacs30/bitwarden-alfred-workfloww#installation), you **_MUST_** add Alfred to the list of Developer Tool exceptions for Alfred to run any workflow that contains an executable (like this one)

