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
    "regexp"
    "strings"
)

// ParseOSReleaseFile parses `/etc/os-release` files
// and returns a map with its key=value's
func ParseOSReleaseFile() {
    b, err := ioutil.ReadFile("/etc/os-release")
    if err != nil {
        fmt.Print(err)
    }
    str := string(b)
    lines := strings.Split(str, "\n")

    props := make(map[string]string)

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
                    // TODO1: Handle comma-type codename string parsing
                    props["codename"] = codename
                } else {
                    props["codename"] = ""
                }
            }
            props[strings.ToLower(kv[0])] = strings.Trim(kv[1], "\"")
        }
    }
    for k, v := range props {
        fmt.Printf("key[%s] value[%s]\n", k, v)
    }
}

// ParseLSBRelease parses the contents of the `lsb_release -a` command
// and returns a map with its key=value's
func ParseLSBRelease() {
    props := make(map[string]string)

    var (
        cmdOut []byte
        err    error
    )
    cmdName := "lsb_release"
    cmdArgs := []string{"-a"}
    if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
        fmt.Fprintln(os.Stderr, "Failed to run lsb_release -a", err)
        os.Exit(1)
    }
    str := string(cmdOut)
    lines := strings.Split(str, "\n")

    for _, element := range lines {
        trimmedElement := strings.Trim(element, "\n")
        if strings.Contains(trimmedElement, "=") {
            kv := strings.Split(trimmedElement, ":")
            key := strings.Replace(kv[0], " ", "_", -1)
            key = strings.ToLower(key)
            props[key] = strings.TrimSpace(kv[1])
        }
    }
}

// ParseDistroReleaseFile parses a distro-specific release/version
// file and returns a map of its data. Not all data is necessarily
// found in each release file and that depends on the distribution
func ParseDistroReleaseFile() {
    b, err := ioutil.ReadFile("/etc/os-release")
    if err != nil {
        fmt.Print(err)
    }
    str := string(b)
    lines := strings.Split(str, "\n")

    props := make(map[string]string)

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
                    // TODO1: Handle comma-type codename string parsing
                    props["codename"] = codename
                } else {
                    props["codename"] = ""
                }
            }
            props[strings.ToLower(kv[0])] = strings.Trim(kv[1], "\"")
        }
    }
    for k, v := range props {
        fmt.Printf("key[%s] value[%s]\n", k, v)
    }
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
    // ParseOSReleaseFile()
    // ParseLSBRelease()
}
