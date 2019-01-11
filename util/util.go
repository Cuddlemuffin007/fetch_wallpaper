
package util

import (
    "fmt"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strconv"
    "strings"

    "../web"
)


func HandleError(err error) {
    if re, ok := err.(*web.RequestError); ok {
        fmt.Fprintf(os.Stderr, "%s", re.Message)
        os.Exit(re.Code)
    } else {
        // some unexpected error must have occurred
        panic(err)
    }
}


func ExpandPath(path string) (string, error) {
    usr, err := user.Current()
    if err != nil {
        return "", nil
    }
    dir := usr.HomeDir
    if path == "~" {
        path = dir
    } else if strings.HasPrefix(path, "~/") {
        path = filepath.Join(dir, path[2:])
    }
    return path, nil
}


func SetBackgroundMacOS(filePath string) error {
    return exec.Command("osascript", "-e", `tell application "Finder" to set desktop picture to POSIX file ` + strconv.Quote(filePath)).Run()
}


// not a full solution because this will (probably) only work on Ubuntu since that what I primarily use.
// TODO: add support for other common distros
func SetBackgroundLinux(filePath string) error {
    return exec.Command("gsettings", "set", "org.gnome.desktop.background", "picture-uri", strconv.Quote("file://" + filePath)).Run()
}
