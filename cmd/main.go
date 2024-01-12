package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/heilkit/tt/tt"
	"log"
	"os"
	"strings"
	"time"
)

const MB = 1 << 20

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
	maxSize := flag.Int("max-size", 1<<10, "download only videos smaller than <VALUE> MB")
	json_ := flag.Bool("json", false, "print info as json, don't download")
	debug := flag.Bool("debug", false, "log debug info")
	quiet_ := flag.Bool("quiet", false, "quiet")
	flag.Parse()

	tt.Debug = *debug
	urls := flag.Args()
	if len(urls) == 0 {
		println("no arguments were passed, use -help to get help")
		os.Exit(0)
	}
	quiet = *quiet_

	for _, url := range urls {
		ensureDir(*directory)

		switch {
		case *cmdProfile:
			until, err := time.Parse(time.DateTime, *until)
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}

			postChan, expectedCount, err := tt.GetUserFeed(url, tt.FeedOpt{
				While: tt.WhileAfter(until),
				OnError: func(err error) {
					if err != nil {
						log.Fatalf("%s: %s", url, err.Error())
					}
				},
				SD: *sd,
				Filter: func(post *tt.Post) bool {
					return post.Size < *maxSize*MB
				},
			})

			jsonRet := []string{}
			i := 0
			for vid := range postChan {
				if *json_ {
					buffer, err := json.MarshalIndent(vid, "", "\t")
					if err != nil {
						log.Fatalf("%s: %s", url, err.Error())
					}
					jsonRet = append(jsonRet, string(buffer))
					continue
				}

				filename, err := vid.Download(tt.DownloadOpt{
					Directory:    *directory,
					ValidateWith: tt.ValidateWithFfprobe(),
				})
				if err != nil {
					log.Fatalf("%s: %s", url, err.Error())
				}
				i += 1
				printf("%s: [%d/%d]\t %s", url, i, expectedCount, filename)
			}

			if *json_ {
				fmt.Printf("[%s]", strings.Join(jsonRet, ",\n"))
			}

		case *cmdInfo:
			vid, err := tt.GetUserDetail(url)
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}

			buffer, err := json.MarshalIndent(vid, "", "\t")
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}
			print(string(buffer))

		default:
			vid, err := tt.GetPost(url, !*sd)
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}

			if *json_ {
				buffer, err := json.MarshalIndent(vid, "", "\t")
				if err != nil {
					log.Fatalf("%s: %s", url, err.Error())
				}
				print(string(buffer))
				continue
			}

			filename, err := vid.Download(
				tt.DownloadOpt{
					Directory:    *directory,
					ValidateWith: tt.ValidateWithFfprobe(),
				})
			if err != nil {
				log.Fatalf("%s: %s", url, err.Error())
			}
			printf("%s: %s", url, filename)
		}

	}
}

var quiet = false

func printf(format string, what ...interface{}) {
	if quiet {
		return
	}
	log.Printf(format, what...)
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, 0755); err != nil {
			log.Fatalln(err)
		}
	}
}
