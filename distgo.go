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

func readFileContents(filePath string) string {
    contentBytes, err := ioutil.ReadFile(filePath)
    if err != nil {
        fmt.Print(err)
    }
    return string(contentBytes)
}

// GetOSReleaseFileInfo retrieves parsed information from an
// os-release file and returns a map with its key-value's
func GetOSReleaseFileInfo(shouldPrint bool) map[string]string {

    defaultMap := make(map[string]string)

    osReleaseFilePath := path.Join(unixEtcDir, osReleaseFileName)
    if _, err := os.Stat(osReleaseFilePath); err == nil {
        content := readFileContents(osReleaseFilePath)
        parsedContent := parseOSReleaseFile(content)
        if shouldPrint {
            for k, v := range parsedContent {
                fmt.Printf("key[%s] value[%s]\n", k, v)
            }
        }
        return parsedContent
    }
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

// GetLSBReleaseInfo retrieves parsed information from an
// `lsb_release -a` command and returns a map with its key-value's
func GetLSBReleaseInfo() map[string]string {
    defaultMap := make(map[string]string)

    var (
        cmdOut []byte
        err    error
    )
    cmdName := "lsb_release"
    cmdArgs := []string{"-a"}
    if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
        fmt.Fprintln(os.Stderr, "Failed to run lsb_release -a", err)
        return defaultMap
    }
    content := string(cmdOut)
    return parseLSBRelease(content)
}

// ParseLSBRelease parses the contents of the `lsb_release -a` command
// and returns a map with its key=value's
func parseLSBRelease(content string) map[string]string {
    props := make(map[string]string)
    lines := strings.Split(content, "\n")

    for _, element := range lines {
        trimmedElement := strings.Trim(element, "\n")
        if strings.Contains(trimmedElement, "=") {
            kv := strings.Split(trimmedElement, ":")
            key := strings.Replace(kv[0], " ", "_", -1)
            key = strings.ToLower(key)
            props[key] = strings.TrimSpace(kv[1])
        }
    }
    return props
}

func stringInSlice(str string, list []string) bool {
    for _, element := range list {
        if element == str {
            return true
        }
    }
    return false
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
    validID := regexp.MustCompile(distroFileNamePattern)

    files, _ := ioutil.ReadDir(unixEtcDir)
    for _, f := range files {
        isReleaseFile := validID.MatchString(f.Name())
        if isReleaseFile {
            releaseFilePath := path.Join(unixEtcDir, f.Name())
            if !stringInSlice(f.Name(), ignoredBasenames) {
                fmt.Println(releaseFilePath)
                content := readFileContents(releaseFilePath)
                parseDistroReleaseFile(content)
            }
        }
    }
    return defaultMap
}

// Reverse returns its argument string reversed rune-wise left to right.
func Reverse(s string) string {
    r := []rune(s)
    for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
        r[i], r[j] = r[j], r[i]
    }
    return string(r)
}

// ParseDistroReleaseFile parses a distro-specific release/version
// file and returns a map of its data. Not all data is necessarily
// found in each release file and that depends on the distribution
func parseDistroReleaseFile(content string) map[string]string {
    props := make(map[string]string)
    line := strings.Split(content, "\n")[0]

    distroFileContentReversePattern := `(?:[^)]*\)(.*)\()? *(?:STL )?([\d.+\-a-z]*\d) *(?:esaeler *)?(.+)`
    compiledPattern := regexp.MustCompile(distroFileContentReversePattern)
    matches := compiledPattern.FindAllString(Reverse(line), -1)
    // props["codename"] = Reverse(matches[1])
    // props["version_id"] = Reverse(matches[2])
    // props["name"] = Reverse(matches[3])
    fmt.Println(matches[1])
    return props
    // matches = _DISTRO_RELEASE_CONTENT_REVERSED_PATTERN.match(
    //     line.strip()[::-1])
    // distro_info = {}
    // if matches:
    //     # regexp ensures non-None
    //     distro_info['name'] = matches.group(3)[::-1]
    //     if matches.group(2):
    //         distro_info['version_id'] = matches.group(2)[::-1]
    //     if matches.group(1):
    //         distro_info['codename'] = matches.group(1)[::-1]
    // elif line:
    //     distro_info['name'] = line.strip()
    // return distro_info

    // basenames = os.listdir(_UNIXCONFDIR)
    // // We sort for repeatability in cases where there are multiple
    // // distro specific files; e.g. CentOS, Oracle, Enterprise all
    // // containing `redhat-release` on top of their own.
    // basenames.sort()
    // for basename in basenames:
    //     if basename in _DISTRO_RELEASE_IGNORE_BASENAMES:
    //         continue
    //     match = _DISTRO_RELEASE_BASENAME_PATTERN.match(basename)
    //     if match:
    //         filepath = os.path.join(_UNIXCONFDIR, basename)
    //         distro_info = self._parse_distro_release_file(filepath)
    //         if 'name' in distro_info:
    //             # The name is always present if the pattern matches
    //             self.distro_release_file = filepath
    //             distro_info['id'] = match.group(1)
    //             return distro_info
    // return {}
}

func main() {
    // GetOSReleaseFileInfo(true)
    // ParseLSBRelease()
    GetDistroReleaseFileInfo()
}
