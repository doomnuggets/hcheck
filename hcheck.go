package main

import (
    "path/filepath"
    "bufio"
    "io"
    "os"
    "fmt"
    "log"
    "crypto/sha256"
    "strings"
    flag "github.com/spf13/pflag"
)

func sha256File(path string) string {
    fh, err := os.Open(path)
    if err != nil {
        log.Fatal(err)
    }
    defer fh.Close()
    h := sha256.New()
    if _, err := io.Copy(h, fh); err != nil {
        log.Fatal(err)
    }
    return fmt.Sprintf("%x", h.Sum(nil))
}


// Stolen from https://gist.github.com/sethamclean/9475737
func fileWalk(location string) (chan string) {
    channel := make(chan string)
    go func() {
        filepath.Walk(location, func(path string, finfo os.FileInfo, _ error)(err error){
            if !finfo.IsDir() {
                channel <- path
            }
            return
        })
        defer close(channel)
    }()
    return channel
}

// Check if a given hash (needle) exists in an array of [filename, hash] (haystack)
func hashMismatch(hashdigest string, filename string, haystack map[string]string) bool {
    return haystack[filename] != hashdigest && haystack[filename] != ""
}

// Check if a given hash and filename exists in an array of [filename, hash] (haystack)
func hashdigestExistsAndMatches(hashdigest string, filename string, haystack map[string]string) bool {
    return haystack[filename] == hashdigest
}


// Parse the sha256sum file into a map of map[filename] = hash
func parseHashfile(hashFilepath string) map[string]string {
    file, err := os.Open(hashFilepath)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    hashMap := make(map[string]string)
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        splittedLine := strings.Split(scanner.Text(), "  ")
        if len(splittedLine) == 2 {
            theHash := splittedLine[0]
            theFilename := splittedLine[1]
            hashMap[theFilename] = theHash
        }
    }
    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

    return hashMap
}

func main() {
    var hashFile = flag.StringP("hash-file", "f", "", "(required) List of hashes to check against.")
    var checkDir = flag.StringP("check-dir", "c", "", "(required) Directory which is scanned and compared against hashes in the hash file.")
    var excludeOK = flag.BoolP("exclude-ok", "o", false, "Exclude status OK lines. (default: false)")
    var excludeMISMATCH = flag.BoolP("exclude-mismatch", "m", false, "Exclude status MISMATCH lines. (default: false)")
    var excludeREMOVED = flag.BoolP("exclude-removed", "r", false, "Exclude status REMOVED lines. (default: false)")
    var excludeNEW = flag.BoolP("exclude-new", "n", false, "Exclude status NEW lines. (default: false)")
    flag.Parse()

    if *excludeOK && *excludeMISMATCH && *excludeREMOVED && *excludeNEW {
        return
    }

    if *hashFile == "" && *checkDir == "" {
        flag.Usage()
        return
    }
    if _, err := os.Open(*hashFile); err != nil {
        log.Print("Unable to open --hash-file=" + *hashFile)
        log.Fatal(err)
    }
    if info, err := os.Stat(*checkDir); err != nil {
        log.Print("Unable to scan --check-dir=" + *checkDir)
        log.Fatal(err)
    } else {
        if info.Mode().Perm() & (1 << (uint(7))) < 5 {
            log.Fatal("Permission denied while accessing --check-dir=" + *checkDir)
        }
    }

    if hashFile != nil && checkDir != nil {
        hashes := parseHashfile(*hashFile)
        for filename := range fileWalk(*checkDir) {
            h := sha256File(filename)
            if hashdigestExistsAndMatches(h, filename, hashes) {
                if !*excludeOK { fmt.Printf("%s  %s: OK\n", h, filename) }
            } else if hashMismatch(h, filename, hashes) {
                if !*excludeMISMATCH { fmt.Printf("%s  %s: MISMATCH\n", h, filename) }
            } else {
                if !*excludeNEW { fmt.Printf("%s  %s: NEW\n", h, filename) }
            }
        }
        // Check for hash file entry files which are missing on the filesystem.
        if !*excludeREMOVED {
            for filename, h := range hashes {
                if _, err := os.Stat(filename); os.IsNotExist(err) {
                    fmt.Printf("%s  %s: REMOVED\n", h, filename)
                }
            }
        }
    }
}
