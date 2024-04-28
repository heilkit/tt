package main

import (
	"flag"
	"fmt"
	"github.com/heilkit/tt/tt"
	"log/slog"
	"os"
)

func main() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-profile | -info] [args...] <urls | usernames | ids>\n", os.Args[0])
		flag.PrintDefaults()
	}
	cmdProfile := flag.Bool("profile", false, "download/scan profiles")
	cmdInfo := flag.Bool("info", false, "print info about profiles")
	until := flag.String("until", "1970-01-01 00:00:00", "don't download videos earlier than")
	sd := flag.Bool("sd", false, "don't request HD sources of videos (less requests => notably faster)")
	directory := flag.String("dir", "./", "directory to save files")
	to_ := flag.String("to", "", "filename to save the video (the default is generated automatically)")
	maxSize := flag.Int64("max-size", 4096, "download only videos smaller than <VALUE> MB")
	retries := flag.Int("retries", 3, "retries number, if something goes wrong")
	json_ := flag.Bool("json", false, "print info as json, don't download")
	debug := flag.Bool("debug", false, "log debug info")
	quiet_ := flag.Bool("quiet", false, "print only errors")
	ignore := flag.Bool("ignore", false, "ignore errors and continue downloading")
	flag.Parse()
	if *retries == 0 {
		*retries = -1
	}

	tt.Debug = *debug
	urls := flag.Args()
	if len(urls) == 0 {
		println("no arguments were passed, use -help to get help")
		os.Exit(0)
	}

	log = slog.Default()
	if *quiet_ || *debug {
		log = slog.New(slog.NewTextHandler(os.Stdout, getOptions(*debug, *quiet_)))
	}
	if *json_ {
		log = slog.New(slog.NewJSONHandler(os.Stdout, getOptions(*debug, *quiet_)))
	}

	for _, url := range urls {
		ensureDir(*directory)

		switch {
		case *cmdProfile:
			if err := CmdProfile(url, CmdProfileOpt{
				SD:        *sd,
				json:      *json_,
				until:     *until,
				retries:   *retries,
				maxSize:   *maxSize,
				directory: *directory,
				ignore:    *ignore,
			}); err != nil {
				log.Error("Downloading profile failed", "user", url, "error", err)
			}

		case *cmdInfo:
			CmdInfo(url)

		default:
			CmdVideo(url, sd, json_, to_, directory, retries)
		}

	}
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			log.Error("error creating directory", "error", err)
		}
	}
}
