package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	flag "github.com/spf13/pflag"
	"golang.org/x/sys/unix"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
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
func fileWalk(location string) chan string {
	channel := make(chan string)
	go func() {
		filepath.Walk(location, func(path string, finfo os.FileInfo, walkErr error) (err error) {
			if walkErr != nil {
				log.Fatal(walkErr)
			}
			if !finfo.IsDir() {
				channel <- path
			}
			return
		})
		defer close(channel)
	}()
	return channel
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

func dirReadable(dir string) bool {
	return unix.Access(dir, unix.R_OK|unix.X_OK) == nil
}

func checksumOk(filename string, hash string, hashes map[string]string) bool {
	return hashes[filename] == hash
}

func checksumMismatch(filename string, hash string, hashes map[string]string) bool {
	return hashes[filename] != hash && hashes[filename] != ""
}

func checksumNew(filename string, hash string, hashes map[string]string) bool {
	return hashes[filename] == ""
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
	if f, err := os.Open(*hashFile); err != nil {
		log.Print("Unable to open --hash-file=" + *hashFile)
		log.Fatal(err)
	} else {
		defer f.Close()
	}

	if !dirReadable(*checkDir) {
		log.Fatal("Unable to traverse --check-dir=" + *checkDir)
		return
	}

	hashes := parseHashfile(*hashFile)
	for filename := range fileWalk(*checkDir) {
		hash := sha256File(filename)
		if checksumOk(filename, hash, hashes) {
			if !*excludeOK {
				fmt.Printf("%s  %s: OK\n", hash, filename)
			}
		} else if checksumMismatch(filename, hash, hashes) {
			if !*excludeMISMATCH {
				fmt.Printf("%s  %s: MISMATCH\n", hash, filename)
			}
		} else if checksumNew(filename, hash, hashes) {
			if !*excludeNEW {
				fmt.Printf("%s  %s: NEW\n", hash, filename)
			}
		} else {
			panic("How the hell did we end up here?")
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
