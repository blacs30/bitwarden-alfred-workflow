package alfred

import (
    "encoding/json"
    "fmt"
    aw "github.com/deanishe/awgo"
    "github.com/jychri/tilde"
    "io/ioutil"
    "log"
    "os"
    "strings"
)

const (
    server                    = "SERVER_URL"
    email                     = "EMAIL"
    sfa                       = "2FA_ENABLED"
    sfaMode                   = "2FA_MODE"
    BwauthKeyword             = "bwauth_keyword"
    BwconfKeyword             = "bwconf_keyword"
    BwKeyword                 = "bw_keyword"
    BwExec                    = "BW_EXEC"
    mod1                      = "MODIFIER_1"
    mod2                      = "MODIFIER_2"
    mod3                      = "MODIFIER_3"
    mod4                      = "MODIFIER_4"
    ReorderingDisabled        = "REORDERING_DISABLED"
    MaxResults                = "MAX_RESULTS"
    EmptyDetailResults        = "EMPTY_DETAIL_RESULTS"
    IconCacheEnabled          = "ICON_CACHE_ENABLED"
    IconCacheTimeout          = "ICON_CACHE_TIMEOUT"
    AutoFetchIconCacheTimeout = "AUTO_FETCH_ICON_CACHE_TIMEOUT"
    SyncCacheTimeout          = "SYNC_CACHE_TIMEOUT"
    CacheTimeout              = "CACHE_TIMEOUT"
    OutputFolder              = "OUTPUT_FOLDER"
)

// Get keys
func GetBwConfKeyword(wf *aw.Workflow) string {
    return wf.Config.GetString(BwconfKeyword, ".bwconfig")
}

func GetCacheTimeout(wf *aw.Workflow) int {
    return wf.Config.GetInt(CacheTimeout, 1440)
}

func GetOutputFolder(wf *aw.Workflow) string {
    folder := wf.Config.GetString(OutputFolder, fmt.Sprintf("%s/Downloads/", ""))
    if folder == "" {
        defaultFolder, err := os.UserHomeDir()
        if err != nil {
            log.Print("Error getting UserHomeDir ", err)
        }
        folder = fmt.Sprintf("%s/Downloads/", defaultFolder)
    } else {
        // tilde.Abs() expands ~ to /Users/$user
        folder = tilde.Abs(folder)
        // in case a "/" exist on the right side, remove it and add it again to be sure it exists.
        folder = fmt.Sprintf("%s/", strings.TrimRight(folder, "/"))
    }
    return folder
}

func GetMaxResults(wf *aw.Workflow) int {
    return wf.Config.GetInt(MaxResults, 20)
}

func GetIconCacheEnabled(wf *aw.Workflow) bool {
    return wf.Config.GetBool(IconCacheEnabled, true)
}

func GetIconCacheTimeout(wf *aw.Workflow) int {
    return wf.Config.GetInt(IconCacheTimeout, 43200)
}

// If an error occurs while getting the icons try to get this icon
// not only after this timeout is expired
func GetAutoFetchIconCacheTimeout(wf *aw.Workflow) int {
    return wf.Config.GetInt(AutoFetchIconCacheTimeout, 1440)
}

func GetSyncCacheTimeout(wf *aw.Workflow) int {
    return wf.Config.GetInt(SyncCacheTimeout, 1440)
}

func GetEmptyDetailResults(wf *aw.Workflow) bool {
    return wf.Config.GetBool(EmptyDetailResults, true)
}
func GetReorderingDisabled(wf *aw.Workflow) bool {
    return wf.Config.GetBool(ReorderingDisabled, true)
}

func GetBwExec(wf *aw.Workflow) string {
    return wf.Config.GetString(BwExec, "bw")
}

func GetBwauthKeyword(wf *aw.Workflow) string {
    return wf.Config.GetString(BwauthKeyword, ".bwauth")
}

func GetBwKeyword(wf *aw.Workflow) string {
    return wf.Config.GetString(BwKeyword, ".bw")
}

func GetServer(wf *aw.Workflow) string {
    return wf.Config.GetString(server)
}

func GetEmail(wf *aw.Workflow) string {
    getEmail := wf.Config.GetString(email, "")
    if getEmail == "" {
        var bwData BwData
        succ, err := OpenBitwardenData(&bwData)
        if err != nil {
            log.Println(err)
            return ""
        }
        if succ {
            err := SetEmail(wf, bwData.UserEmail)
            if err != nil {
               log.Println(err)
               return ""
            }
            getEmail = bwData.UserEmail
        }
    }
    return getEmail
}

func GetSfa(wf *aw.Workflow) bool {
    return wf.Config.GetBool(sfa, true)
}

func GetSfaMode(wf *aw.Workflow) int {
    return wf.Config.GetInt(sfaMode, 0)
}

//// Modifiers
func GetMod1(wf *aw.Workflow) string {
    return wf.Config.GetString(mod1, "alt")
}
func GetMod2(wf *aw.Workflow) string {
    return wf.Config.GetString(mod2, "shift")
}
func GetMod3(wf *aw.Workflow) string {
    return wf.Config.GetString(mod3, "cmd")
}
func GetMod4(wf *aw.Workflow) string {
    return wf.Config.GetString(mod4, "cmd,alt,ctrl")
}

// Set keys
func SetServer(wf *aw.Workflow, url string) error {
    return wf.Config.Set(server, url, false).Do()
}

func SetEmail(wf *aw.Workflow, address string) error {
    return wf.Config.Set(email, address, false).Do()
}

func SetSfa(wf *aw.Workflow, enabled string) error {
    return wf.Config.Set(sfa, enabled, true).Do()
}

func SetSfaMode(wf *aw.Workflow, id string) error {
    return wf.Config.Set(sfaMode, id, true).Do()
}

func OpenBitwardenData(bwData interface{}) (bool, error) {
    homedir, err := os.UserHomeDir()
    if err != nil {
        return false, err
    }
    bwDataPath := fmt.Sprintf("%s/Library/Application Support/Bitwarden CLI/data.json", homedir )
    log.Println("BW DataPath", bwDataPath)
    if _, err := os.Stat(bwDataPath); err != nil {
        log.Println("Couldn't find the Bitwarden data.json ", err)
        return false, err
    }
    data, err := ioutil.ReadFile(bwDataPath)
    if err != nil {
        return false, err
    }
    if err := json.Unmarshal(data, &bwData); err != nil {
        log.Printf("Couldn't load the items cache, error: %s", err)
        return false, err
    }
    log.Println("Got existing Bitwarden CLI data")
    return true, nil
}

type BwData struct {
    InstalledVersion string                 `json:"installedVersion"`
    UserEmail        string                 `json:"userEmail"`
    Unused           map[string]interface{} `json:"-"`
}
