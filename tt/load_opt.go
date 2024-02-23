package tt

import (
	"encoding/json"
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"os/exec"
	"time"
)

type DownloadOpt struct {
	Directory      string
	To             string
	DownloadWith   func(url string, filename string) error
	ValidateWith   func(filename string) (bool, error)
	FilenameFormat func(post *Post, i int) string
	Fallback       func(post *Post, opt DownloadOpt, err error) (files []string, e error)
	Timeout        time.Duration
	TimeoutOnError time.Duration
	NoSync         bool
	Retries        int
}

func (opt *DownloadOpt) Defaults() *DownloadOpt {
	ret := opt
	if ret == nil {
		ret = &DownloadOpt{}
	}
	if ret.DownloadWith == nil {
		ret.DownloadWith = func(url string, filename string) error {
			req, err := grab.NewRequest(filename, url)
			if err != nil {
				return err
			}

			if resp := DefaultDownloadClient.Do(req); resp.Err() != nil {
				return err
			}
			return nil
		}
	}
	if ret.ValidateWith == nil {
		ret.ValidateWith = func(filename string) (bool, error) {
			return true, nil
		}
	}
	if ret.FilenameFormat == nil {
		ret.FilenameFormat = formatFilename
	}
	if ret.Timeout < 0 {
		ret.Timeout = 0
	} else if ret.Timeout == 0 {
		ret.Timeout = DefaultDownloadTimeout
	}
	if ret.TimeoutOnError < 0 {
		ret.TimeoutOnError = 0
	} else if ret.TimeoutOnError == 0 {
		ret.TimeoutOnError = DefaultDownloadTimeoutOnError
	}
	if ret.Retries < 0 {
		ret.Retries = 0
	} else if ret.Retries == 0 {
		ret.Retries = DefaultRetries
	}
	if ret.Fallback == nil {
		ret.Fallback = FallbackNone
	}
	return ret
}

func FallbackNone(post *Post, opt DownloadOpt, err error) (files []string, e error) {
	return nil, err
}

func FallbackSD(post *Post, opt DownloadOpt, err error) (files []string, e error) {
	post, err = GetPost(post.Id, false)
	if err != nil {
		return nil, fmt.Errorf("falling back in SD failed with %s", err.Error())
	}
	opt.Fallback = FallbackNone
	return post.Download(opt)
}

func ValidateWithFfprobe(ffprobe ...string) func(filename string) (isValid bool, err error) {
	ffprobe_ := "ffprobe"
	if len(ffprobe) != 0 {
		ffprobe_ = ffprobe[0]
	}

	return func(filename string) (bool, error) {
		out, err := exec.Command(ffprobe_, "-loglevel", "error", "-of", "json", "-show_entries", "stream_tags:format_tags", filename).CombinedOutput()
		if err != nil {
			return false, fmt.Errorf("err: %s,\n%s", err.Error(), string(out))
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(out, &metadata); err != nil || len(metadata) == 0 {
			return false, fmt.Errorf("err: %s,\n%s", err.Error(), string(out))
		}

		return true, nil
	}
}
