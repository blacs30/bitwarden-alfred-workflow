# coding=utf-8
import json
from subprocess import Popen, PIPE
import os

search = "{query}"
FOLDER_SWITCH = '-f'

my_env = os.environ.copy()
my_env["PATH"] = "/usr/local/bin:/opt/local/bin:" + my_env["PATH"]

result = {"items": []}  # type: dict


class SearchType(object):
    items = 'items'
    folders = 'folders'


def get_session_key():
    cmd = 'security find-generic-password -w -s alfred-bitwarden-session-key'
    proc = Popen(cmd.split(), stdout=PIPE)
    output = proc.stdout.read().decode('utf8')
    if not output:
        print(json.dumps(error_result(type="login")))
        exit(1)
    return output


def check_login():
    proc = Popen("launchctl getenv BW_ASKPASS".split(), stdout=PIPE)
    output = proc.stdout.read().decode('utf8')
    if not output:
        print(json.dumps(error_result(type="locked")))
        exit(1)
    return


def add_result(uid, title, subtitle, arg, result_list):
    item = {
        "uid": uid,
        "title": title,
        "subtitle": subtitle,
        "arg": arg,
        "icon": {
            "path": "icon.png"
        }
    }
    result_list["items"].append(item)
    return result_list


def error_result(type):
    uid_switcher = {
        "login": {"uid": "error-login", "title": "It appears you are not logged in to Bitwarden.",
                  "subtitle": "Please login using the \'bwlogin\' command or press \'ctrl\' + enter to login now.",
                  "arg": "scriptlocationnotset"},
        "install": {"uid": "error-noinstall", "title": "You do not have the Bitwarden CLI Installed.",
                    "subtitle": "Press enter to be taken to the install instructions.", "arg": "\'error-noinstall\'"},
        "none": {"uid": "error-nonefount", "title": "No search results matching your query found.",
                 "subtitle": "Please try again with a different query.", "arg": "\'error-nonefound\'"},
        "locked": {"uid": "error-locked", "title": "Your vault is locked.",
                   "subtitle": "Please unlock your vault with \'bwunlock\' or press \'fn\' + enter to unlock now..",
                   "arg": "scriptlocationnotset"},
    }
    uid = uid_switcher.get(type)["uid"]
    title = uid_switcher.get(type)["title"]
    subtitle = uid_switcher.get(type)["subtitle"]
    arg = uid_switcher.get(type)["arg"]
    output = add_result(uid=uid, title=title, subtitle=subtitle, arg=arg, result_list=result)
    return output


def get_bw_exec():
    bw_exec = ""
    for f in ['/usr/local/bin/bw', '/opt/local/bin/bw', '/usr/bin/bw']:
        if os.path.exists(f):
            bw_exec = f
    if not bw_exec:
        print(json.dumps(error_result(type="install")))
        exit(2)
    return bw_exec


def get_bw_search_result(bw_exec, session_key, bw_list_obj, bw_search, bw_folder_id=None):
    cmd = "{bw_exec} --session={session_key} list {bw_list_obj} --search={bw_search}".format(bw_exec=bw_exec,
                                                                                             session_key=session_key,
                                                                                             bw_list_obj=bw_list_obj,
                                                                                             bw_search=bw_search)
    if bw_folder_id is not None:
        cmd = "{cmd} --folderid={folderid}".format(cmd=cmd, folderid=bw_folder_id)
    proc = Popen(cmd.split(), env=my_env, stdout=PIPE, stderr=PIPE)
    output, err = proc.communicate()
    if "mac failed." in err.decode('utf8').strip():
        print(json.dumps(error_result(type="locked")))
        exit(1)
    if "Vault is locked." in output.decode('utf8').strip():
        print(json.dumps(error_result(type="locked")))
        exit(1)
    try:
        results = json.loads(output)
    except ValueError:
        if "You are not logged in" in output.decode('utf8'):
            print(json.dumps(error_result(type="login")))
            exit(1)
    try:
        if not results:
            print(json.dumps(error_result(type="none")))
            exit(2)
    except NameError:
        print(json.dumps(error_result(type="none")))
        exit(2)
    return results


def parse_search_string(bw_exec, session_key, search):
    """
    Parse user input and split into search string and folder name.
    Then search for folder id by name and return *first* folder id, which is found.

    :param bw_exec: path to Bitwarden executable
    :param session_key: Bitwarden session key
    :param search: user search input
    :return: tuple(search string, folder_id)
    """
    tokens = search.split()
    if FOLDER_SWITCH not in tokens:
        return search, None

    folder_id = None
    token_id = tokens.index(FOLDER_SWITCH)
    try:
        tokens.pop(token_id)
        folder_name = tokens.pop(token_id)
        res = get_bw_search_result(bw_exec=bw_exec, session_key=session_key,
                                   bw_list_obj=SearchType.folders,
                                   bw_search=folder_name)
        folder_id = res[0].get('id')
    except IndexError:
        pass
    finally:
        return ' '.join(tokens), folder_id


def bw_search(bw_exec, session_key, search):
    search_str, folder_id = parse_search_string(bw_exec, session_key, search)
    return get_bw_search_result(bw_exec=bw_exec, session_key=session_key,
                                bw_list_obj=SearchType.items,
                                bw_search=search_str,
                                bw_folder_id=folder_id)


def build_bw_result(bw_exec, session_key, search_result):
    for bw_item in search_result:
        subtitle = ""
        item_id = bw_item["id"]
        name = bw_item["name"]
        try:
            username = bw_item["login"]["username"].encode('utf-8').strip()
        except (KeyError, AttributeError):
            username = "∅"
        finally:
            subtitle += 'Username: {username}'.format(username=username)
        try:
            if bw_item["login"]["totp"]:
                subtitle += " ⸰ OTP: *"
        except (KeyError, AttributeError):
            pass
        add_result(uid=item_id, title=name, subtitle=subtitle, arg=[name, item_id], result_list=result)


session_key = get_session_key()
check_login = check_login()
bw_exec = get_bw_exec()
bw_search_result = bw_search(bw_exec=bw_exec, session_key=session_key, search=search)
build_bw_result(bw_exec=bw_exec, session_key=session_key, search_result=bw_search_result)

print(json.dumps(result))
