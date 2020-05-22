package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

/* Globals */
const magic = "oversized-v001"
const blockLen = 4096 // should be able to store magic, a sha256 and a file length

var logger *log.Logger
var config struct {
	gitPath   string
	gitRoot   string
	objDir    string
	tmpDir    string
	s3Bucket  string
	s3Prefix  string
	s3Profile string
}

type CleanFile struct {
	Magic  string `json:"magic"`
	Sha256 string `json:"sha256"`
	Length int    `json:"length"`
}

/* Magic Init Function */
func init() {
	logger = log.New(os.Stderr, "", log.Lshortfile)
	if bts, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err != nil {
		// not in a git repo
		logger.Printf("Not in a git repository! (%v)\n", err)
		os.Exit(-1)
	} else {
		config.gitRoot = strings.TrimSpace(string(bts))
	}

	if bts, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err != nil {
		// not in a git repo
		logger.Printf("Not in a git repository! (%v)\n", err)
		os.Exit(-1)
	} else {
		config.gitPath = strings.TrimSpace(string(bts))
	}

	config.objDir = path.Join(config.gitPath, "fat", "objects")
	if err := os.MkdirAll(config.objDir, 0755); err != nil {
		// not in a git repo
		logger.Printf("Failed to create directory: %v, %v\n", config.objDir, err)
		os.Exit(-1)
	}

	config.tmpDir = path.Join(config.gitPath, "fat", "tmp")
	if err := os.MkdirAll(config.tmpDir, 0755); err != nil {
		// not in a git repo
		logger.Printf("Failed to create directory: %v, %v\n", config.tmpDir, err)
		os.Exit(-1)
	}

	// pull config, if possible
	if val, err := exec.Command("git", "config", "--get", "bucket").Output(); err != nil {
		config.s3Bucket = string(val)
	}
	if val, err := exec.Command("git", "config", "--get", "prefix").Output(); err != nil {
		config.s3Prefix = string(val)
	}
	if val, err := exec.Command("git", "config", "--get", "profile").Output(); err != nil {
		config.s3Profile = string(val)
	}
}

/* Helpers */
func printUsage() {
	logger.Println("Usage git phat [init|status|push|pull|gc|verify|checkout|find|index-filter|filter-clean|filter-smudge]")
}

func filterClean(istream io.Reader, ostream io.Writer) {
	// convert from smudged to magic / clean version
	var err error
	var tmpFile *os.File
	hasher := sha256.New()

	// we're in the process of receiving a large file, write to temp file then
	// copy it to object dir
	if tmpFile, err = ioutil.TempFile(config.tmpDir, "oversized-*"); err != nil {
		logger.Printf("Failed to create temporary directory: %v", err)
		os.Exit(-1)
	}
	defer os.Remove(tmpFile.Name())

	// Hash and copy to temp
	var writers []io.Writer
	writers = append(writers, tmpFile)
	writers = append(writers, hasher)
	dest := io.MultiWriter(writers...)

	var fullSize int64
	if fullSize, err = io.Copy(dest, istream); err != nil {
		logger.Printf("Failed to copy stream: %v", err)
		os.Exit(-1)
	}

	// Read a block of tmpfile
	block := make([]byte, blockLen)
	tmpFile.Seek(0, 0)
	var nBytes int
	if nBytes, err = tmpFile.Read(block); err != nil {
		logger.Printf("Failed to read from tmpfile, %v", err)
		os.Exit(-1)
	}

	var cleanFile CleanFile
	err = json.Unmarshal([]byte(block[0:nBytes]), &cleanFile)
	logger.Printf("Parsing, %v", block)
	if err == nil {
		logger.Printf("Success, contents: %v", cleanFile)
		if cleanFile.Magic == magic {
			// parsed properly and magic matches, so just return the contents of the file (it is already clean)
			logger.Printf("Magic matched, writing out known cleanfile")
			ostream.Write(block)
			return
		}
	} else {
		logger.Printf("Error: %v", err)
	}

	digest := hasher.Sum(nil)
	objFile := path.Join(config.objDir, fmt.Sprintf("%x", digest))
	if _, err := os.Stat(objFile); os.IsNotExist(err) {
		// Set permissions for the new file using the current umask
		if err2 := os.Rename(tmpFile.Name(), objFile); err2 != nil {
			logger.Printf("Error while renaming: %v", err2)
			return
		}
		if err2 := os.Chmod(objFile, 444); err2 != nil {
			logger.Printf("Error while chmoding: %v", err2)
			return
		}
		logger.Printf("git-fat filter-clean: caching to %v", objFile)
	} else {
		logger.Printf("git-fat filter-clean: cache already exists %v", objFile)
	}

	cleanFile.Magic = magic
	cleanFile.Sha256 = fmt.Sprintf("%x", digest)
	cleanFile.Length = int(fullSize)
	if data, err := json.Marshal(cleanFile); err != nil {
		logger.Printf("Failed to marshal: %v", cleanFile)
		os.Exit(-1)
	} else {
		ostream.Write(data)
		ostream.Write([]byte("\n"))
	}
}

func filter_smudge() {}

func initRepo(bucket string, prefix string, profile string) {
	if _, err := exec.Command("git", "config", "oversized.bucket", bucket).Output(); err != nil {
		logger.Printf("Failed to set bucket: %v", err)
	}

	if len(prefix) > 0 {
		if _, err := exec.Command("git", "config", "oversized.prefix", prefix).Output(); err != nil {
			logger.Printf("Failed to set prefix: %v", err)
		}
	}

	if len(profile) > 0 {
		if _, err := exec.Command("git", "config", "oversized.profile", profile).Output(); err != nil {
			logger.Printf("Failed to set profile: %v", err)
		}
	}
}

func status()       {}
func push()         {}
func pull()         {}
func gc()           {}
func verify()       {}
func checkout()     {}
func find()         {}
func index_filter() {}

func main() {
	// first pull from .git/aws_credentials

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
	pushCmd := flag.NewFlagSet("push", flag.ExitOnError)
	pullCmd := flag.NewFlagSet("pull", flag.ExitOnError)
	gcCmd := flag.NewFlagSet("gc", flag.ExitOnError)
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	checkoutCmd := flag.NewFlagSet("checkout", flag.ExitOnError)
	findCmd := flag.NewFlagSet("find", flag.ExitOnError)
	indexFilterCmd := flag.NewFlagSet("index-filter", flag.ExitOnError)
	filterCleanCmd := flag.NewFlagSet("filter-clean", flag.ExitOnError)
	filterSmudgeCmd := flag.NewFlagSet("filter-smudge", flag.ExitOnError)

	//fooEnable := fooCmd.Bool("enable", false, "enable")
	//    fooName := fooCmd.String(
	bucket := initCmd.String("bucket", "", "Bucket to save large file to")
	prefix := initCmd.String("prefix", "", "Path within bucket to save to (otherwise will save to root")
	profile := initCmd.String("profile", "", "AWS profile to use, if pulling credentials from an INI file")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "help":
		printUsage()
	case "init":
		initCmd.Parse(os.Args[2:])
		if *bucket == "" {
			logger.Println("Must provide a bucket!")
			os.Exit(0)
		}
		initRepo(*bucket, *prefix, *profile)
	case "status":
		statusCmd.Parse(os.Args[2:])
	case "push":
		pushCmd.Parse(os.Args[2:])
	case "pull":
		pullCmd.Parse(os.Args[2:])
	case "gc":
		gcCmd.Parse(os.Args[2:])
	case "verify":
		verifyCmd.Parse(os.Args[2:])
	case "checkout":
		checkoutCmd.Parse(os.Args[2:])
	case "find":
		findCmd.Parse(os.Args[2:])
	case "index-filter":
		indexFilterCmd.Parse(os.Args[2:])
	case "filter-clean":
		// The clean filter runs when a file is added to the index. It gets the "smudged" (tree)
		// version of the file on stdin and produces the "clean" (repository) version on stdout.
		filterCleanCmd.Parse(os.Args[2:])
		filterClean(os.Stdin, os.Stdout)
	case "filter-smudge":
		filterSmudgeCmd.Parse(os.Args[2:])
	default:
		printUsage()
		os.Exit(0)
	}
}
