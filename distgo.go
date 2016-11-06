/*
Package disgo implements a simple library for identifying the linux
distribution you're running on.
*/
package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path"
    "regexp"
    "strings"
)

const unixEtcDir string = "/etc"
const osReleaseFileName string = "os-release"

// GetOSReleaseFileInfo retrieves parsed information from an
// os-release file and returns a map with its key-value's
func GetOSReleaseFileInfo() map[string]string {

    defaultMap := make(map[string]string)

    osReleaseFilePath := path.Join(unixEtcDir, osReleaseFileName)
    if _, err := os.Stat(osReleaseFilePath); err == nil {
        content := readFileContents(osReleaseFilePath)
        printMap(parseOSReleaseFile(content))
        return parseOSReleaseFile(content)
    }
    return defaultMap
}

// GetLSBReleaseInfo retrieves parsed information from an
// `lsb_release -a` command and returns a map with its key-value's
func GetLSBReleaseInfo() map[string]string {
    defaultMap := make(map[string]string)
    var (
        cmdOut []byte
        err    error
    )
    cmdName := "/usr/bin/lsb_release"
    cmdArgs := []string{"-a"}

    if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
        fmt.Fprintln(os.Stderr, "Failed to run lsb_release -a", err)
        return defaultMap
    }
    printMap(parseLSBRelease(string(cmdOut)))
    return parseLSBRelease(string(cmdOut))
}

// GetDistroReleaseFileInfo retrieves parsed information from an
// `lsb_release -a` command and returns a map with its key-value's
func GetDistroReleaseFileInfo() map[string]string {
    defaultMap := make(map[string]string)

    ignoredBasenames := []string{
        "debian_version",
        "lsb-release",
        "oem-release",
        "os-release",
        "system-release",
    }

    distroFileNamePattern := `(\w+)[-_](release|version)$`
    compiledPattern := regexp.MustCompile(distroFileNamePattern)

    files, _ := ioutil.ReadDir(unixEtcDir)
    for _, f := range files {
        isReleaseFile := compiledPattern.MatchString(f.Name())
        if isReleaseFile {
            matches := compiledPattern.FindAllStringSubmatch(f.Name(), -1)
            releaseFilePath := path.Join(unixEtcDir, f.Name())
            if !stringInSlice(f.Name(), ignoredBasenames) {
                content := readFileContents(releaseFilePath)
                defaultMap = parseDistroReleaseFile(content)
                if _, ok := defaultMap["name"]; ok {
                    defaultMap["id"] = matches[0][1]
                }
            }
        }
    }
    printMap(defaultMap)
    return defaultMap
}

// ParseOSReleaseFile parses `/etc/os-release` files
// and returns a map with its key=value's
func parseOSReleaseFile(content string) map[string]string {
    props := make(map[string]string)
    lines := strings.Split(content, "\n")

    for _, element := range lines {
        if strings.Contains(element, "=") {
            kv := strings.Split(element, "=")
            if kv[0] == "VERSION" {
                validID := regexp.MustCompile(`(\(\D+\))|,(\s+)?\D+`)
                codenameFound := validID.MatchString(kv[1])
                codename := validID.FindString(kv[1])
                if codenameFound {
                    codename = strings.TrimSpace(codename)
                    codename = strings.Trim(codename, "()")
                    // TODO: Handle comma-type codename string parsing
                    props["codename"] = codename
                } else {
                    props["codename"] = ""
                }
            }
            props[strings.ToLower(kv[0])] = strings.Trim(kv[1], "\"")
        }
    }
    return props
}

// ParseLSBRelease parses the contents of the `lsb_release -a` command
// and returns a map with its key=value's
func parseLSBRelease(content string) map[string]string {
    props := make(map[string]string)
    lines := strings.Split(content, "\n")

    for _, element := range lines {
        trimmedElement := strings.Trim(element, "\n")
        if strings.Contains(trimmedElement, ":") {
            kv := strings.Split(trimmedElement, ":")
            key := strings.Replace(kv[0], " ", "_", -1)
            key = strings.ToLower(key)
            props[key] = strings.TrimSpace(kv[1])
        }
    }
    return props
}

// ParseDistroReleaseFile parses a distro-specific release/version
// file and returns a map of its data. Not all data is necessarily
// found in each release file and that depends on the distribution
func parseDistroReleaseFile(content string) map[string]string {
    props := make(map[string]string)
    line := strings.Split(content, "\n")[0]

    distroFileContentReversePattern := `(?:[^)]*\)(.*)\()? *(?:STL )?([\d.+\-a-z]*\d) *(?:esaeler *)?(.+)`
    compiledPattern := regexp.MustCompile(distroFileContentReversePattern)
    matches := compiledPattern.FindAllStringSubmatch(Reverse(line), -1)
    if len(matches) > 0 {
        groups := matches[0]
        props["name"] = Reverse(groups[3])
        props["version_id"] = Reverse(groups[2])
        props["codename"] = Reverse(groups[1])
    } else if len(line) > 0 {
        props["name"] = strings.TrimSpace(line)
    }
    return props
}

func printMap(content map[string]string) {
    fmt.Println("\n***********************************")
    for k, v := range content {
        fmt.Printf("key[%s] value[%s]\n", k, v)
    }
}

func stringInSlice(str string, list []string) bool {
    for _, element := range list {
        if element == str {
            return true
        }
    }
    return false
}

// Reverse returns its argument string reversed rune-wise left to right.
func Reverse(str string) string {
    r := []rune(str)
    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r)
}

func readFileContents(filePath string) string {
    contentBytes, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Print(err)
    }
    return string(contentBytes)
}

func main() {
    GetOSReleaseFileInfo()
    GetLSBReleaseInfo()
    GetDistroReleaseFileInfo()
}
