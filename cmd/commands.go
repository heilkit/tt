package main

import (
	"encoding/json"
	"fmt"
	"github.com/heilkit/tt/tt"
	"log/slog"
	"strings"
	"time"
)

const unixTimeStart = "1970-01-01 00:00:00"

var log = slog.Default()

const MB = 1 << 20

func CmdInfo(url string) {
	vid, err := tt.GetUserDetail(url)
	if err != nil {
		log.Error(fmt.Sprintf("%s: %s", url, err.Error()))
	}

	buffer, err := json.MarshalIndent(vid, "", "\t")
	if err != nil {
		log.Error(fmt.Sprintf("%s: %s", url, err.Error()))
	}
	print(string(buffer))
}

func CmdVideo(url string, sd *bool, json_ *bool, to_ *string, directory *string, retries *int) {
	post, err := tt.GetPost(url, !*sd)
	if err != nil {
		log.Error(fmt.Sprintf("%s: %s", url, err.Error()))
	}

	if *json_ {
		buffer, err := json.MarshalIndent(post, "", "\t")
		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", url, err.Error()))
		}
		print(string(buffer))

	} else {
		filename, err := post.Download(&tt.DownloadOpt{
			Filename:  *to_,
			Directory: *directory,
			Retries:   *retries,
			Fallback:  tt.FallbackToSD,
			SD:        *sd,
			Log:       log,
		})
		if err != nil {
			log.Error(fmt.Sprintf("%s: %s", url, err.Error()))
		}
		log.Info("Downloaded", "post", post.ID(), "to", filename)
	}
}

type CmdProfileOpt struct {
	SD        bool
	json      bool
	until     string
	retries   int
	maxSize   int64
	directory string
	ignore    bool
}

func CmdProfile(user string, opt CmdProfileOpt) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	log.Info("Starting profile download", "user", user, "HD", !opt.SD)

	until, err := time.Parse(time.DateTime, opt.until)
	if err != nil {
		return fmt.Errorf("could not parse until flag: %w", err)
	}
	if opt.until != unixTimeStart {
		log.Info("Ignoring videos before", "time", opt.until)
	}

	postChan, expectedCount, err := tt.GetUserFeed(user, tt.FeedOpt{
		While: tt.WhileAfter(until),
		OnError: func(err error) {
			if err != nil {
				panic(fmt.Errorf("could not get user feed: %w", err))
			}
		},
		SD:     opt.SD,
		Filter: func(post *tt.Post) bool { return post.Size < opt.maxSize*MB },
	})

	log.Info(fmt.Sprintf("Expecting %d posts", expectedCount))
	jsonList := []string{}
	i := 0
	for post := range postChan {
		i += 1
		if opt.json {
			str, err := json.MarshalIndent(post, "", "  ")
			if err != nil {
				return fmt.Errorf("could not marshal post %s: %w", post.ID(), err)
			}
			jsonList = append(jsonList, string(str))
			continue
		}

		files, err := post.Download(&tt.DownloadOpt{
			Directory: opt.directory,
			Retries:   opt.retries,
			Fallback:  tt.FallbackToSD,
			SD:        opt.SD,
			Log:       log,
		})
		if err != nil {
			err := fmt.Errorf("could not download post %s: %w", post.ID(), err)
			if !opt.ignore {
				return err
			}
			log.Error("While downloading", "post", post.ID(), "err", err)
		}
		log.Info(fmt.Sprintf("[%d/%d]\t Downloaded post %s to %s", i, expectedCount, post.ID(), strings.Join(files, ", ")))
	}

	if opt.json {
		fmt.Printf("[%s]", strings.Join(jsonList, ",\n"))
	}
	log.Info("Download complete", "user", user)

	return nil
}

func getOptions(debug, quiet bool) *slog.HandlerOptions {
	level := slog.LevelInfo
	switch {
	case quiet:
		level = slog.LevelError
	case debug:
		level = slog.LevelDebug
	}
	return &slog.HandlerOptions{
		AddSource: level == slog.LevelDebug,
		Level:     level,
	}
}
