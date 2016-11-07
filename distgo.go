/*
Package distgo implements a simple library for identifying the linux
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

// LinuxDistributionObject ...
type LinuxDistributionObject struct {
    OsReleaseFile     string
    DistroReleaseFile string
    OsReleaseInfo     map[string]string
    LSBReleaseInfo    map[string]string
    DistroReleaseInfo map[string]string
}

// LinuxDistribution ...
func LinuxDistribution(d *LinuxDistributionObject) *LinuxDistributionObject {
    if d == nil {
        d = &LinuxDistributionObject{
            OsReleaseFile: path.Join(unixEtcDir, osReleaseFileName),
        }
    }
    d.OsReleaseInfo = d.GetOSReleaseFileInfo()
    // d.LSBReleaseInfo = d.GetLSBReleaseInfo()
    // d.DistroReleaseInfo = d.GetDistroReleaseFileInfo()
    return d
}

// GetOSReleaseFileInfo retrieves parsed information from an
// os-release file and returns a map with its key-value's
func (d *LinuxDistributionObject) GetOSReleaseFileInfo() map[string]string {
    defaultMap := make(map[string]string)

    if _, err := os.Stat(d.OsReleaseFile); err == nil {
        content := readFileContents(d.OsReleaseFile)
        // printMap(parseOSReleaseFile(content))
        return parseOSReleaseFile(content)
    }
    return defaultMap
}

// GetLSBReleaseInfo retrieves parsed information from an
// `lsb_release -a` command and returns a map with its key-value's
func (d *LinuxDistributionObject) GetLSBReleaseInfo() map[string]string {
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
    // printMap(parseLSBRelease(string(cmdOut)))
    return parseLSBRelease(string(cmdOut))
}

// GetDistroReleaseFileInfo retrieves parsed information from an
// `lsb_release -a` command and returns a map with its key-value's
func (d *LinuxDistributionObject) GetDistroReleaseFileInfo() map[string]string {
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

    // printMap(defaultMap)
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

func (d *LinuxDistributionObject) getOSReleaseAttribute(attribute string) string {
    return d.OsReleaseInfo[attribute]
}

func (d *LinuxDistributionObject) getLSBReleaseAttribute(attribute string) string {
    return d.LSBReleaseInfo[attribute]
}

func (d *LinuxDistributionObject) getDistroReleaseAttribute(attribute string) string {
    return d.DistroReleaseInfo[attribute]
}

// Name returns the name of the distribution
func (d *LinuxDistributionObject) Name(pretty bool) string {
    var name string

    name = d.getOSReleaseAttribute("name")
    if len(name) == 0 {
        name = d.getLSBReleaseAttribute("distributor_id")
    }
    if len(name) == 0 {
        name = d.getDistroReleaseAttribute("name")
    }
    if pretty {
        name = d.getOSReleaseAttribute("pretty_name")
        if len(name) == 0 {
            name = d.getLSBReleaseAttribute("description")
        }
        if len(name) == 0 {
            name = d.getDistroReleaseAttribute("name")
        }
    }
    return name
    //     name = self.os_release_attr('name') \
    //     or self.lsb_release_attr('distributor_id') \
    //     or self.distro_release_attr('name')
    // if pretty:
    //     name = self.os_release_attr('pretty_name') \
    //         or self.lsb_release_attr('description')
    //     if not name:
    //         name = self.distro_release_attr('name')
    //         version = self.version(pretty=True)
    //         if version:
    //             name = name + ' ' + version
    // return name or ''
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
    // Can also pass &LinuxDistributionObject{args} instead
    d := LinuxDistribution(nil)
    fmt.Println(d.Name(true))
}
