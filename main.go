package main

import (
	"flag"
	"fmt"
	"os"
	_ "path"
)

func printUsage() {
	fmt.Println("Usage git phat [init|status|push|pull|gc|verify|checkout|find|index-filter|filter-clean|filter-smudge]")
}

func main() {
	// first pull from .git/aws_credentials
	gitRoot, err := Repo()
	if err != nil {
		// not in a git repo
		fmt.Println("Not in a git repository!")
		os.Exit(-1)
	}

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
		if bucket == nil {
			fmt.Println("Must provide a bucket!")
			os.Exit(0)
		}
		Init(gitRoot.Path, *bucket, *prefix, *profile)
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
		filterCleanCmd.Parse(os.Args[2:])
	case "filter-smudge":
		filterSmudgeCmd.Parse(os.Args[2:])
	default:
		printUsage()
		os.Exit(0)
	}
}
