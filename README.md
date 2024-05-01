# Tikwm API

https://tikwm.com is the best middleman for getting TikTok video info, afaik. If I'm wrong, contact me immediately.

* Download videos in HD
* Download whole profiles in a minimal time (with reasonable naming)

## [Executable] Example

* `./tikmeh "https://www.tiktok.com/@locallygrownwig/video/6901498776523951365"` -- download this video in HD to current
  folder
* `./tikmeh -profile losertron` -- download all @losertron content
* `./tikmeh -profile -until "2023-01-01 00:00:00" losertron` -- download all @losertron content from 2023 to now
* `./tikmeh -info losertron` -- get user info about @losertron profile

```
$ ./tikmeh
Usage: ./tikmeh [-profile | -info] [args...] <urls | usernames | ids>
  -profile
        download/scan profiles
  -info
        print info about profiles
  -dir string
        directory to save files (default "./")
  -to string
        filename to save the video (the default is generated automatically)
  -json
        print info as json, don't download
  -debug
        log debug info
  -quiet
        print only errors
  -max-size int
        download only videos smaller than <VALUE> MB (default 4096)
  -retries int
        retries number, if something goes wrong (default 3)
  -ignore
        ignore errors and continue downloading
  -sd
        don't request HD sources of videos (less requests => notably faster)
  -until string
        don't download videos earlier than (default "1970-01-01 00:00:00")
```

### Download & Setup executable

Go to releases — https://github.com/heilkit/tt/releases. Choose an executable that suits your system and have fun, 
everything you need is packed in already.

## [Library] Example

```go
package main

import (
	"github.com/heilkit/tt/tt"
	"log"
	"time"
)

func main() {
    // basic, simplest way
	postInfo, files, err := tt.Download("https://www.tiktok.com/@locallygrownwig/video/6901498776523951365")

	// tt.GetVideo(url string, HD bool) ()
	postHD, err := tt.GetPost("https://www.tiktok.com/@locallygrownwig/video/6901498776523951365")
	postHD, err = tt.GetPost("6901498776523951365", true)                // with ID
	postSD, err := tt.GetPost("https://vm.tiktok.com/ZM66UoB9m/", false) // with shorten link
	localname, err := postHD.Download(&tt.DownloadOpt{Filename: "locallygrownwig.mp4"})

	// Get user posts for the last 30 days
	until := time.Now().Add(-time.Hour * 24 * 30)
	vidChan, expectedCount, err := tt.GetUserFeed("locallygrownwig", tt.FeedOpt{
		While:  tt.WhileAfter(until),
		Filter: tt.FilterVideo,
	})

	for vid := range vidChan {
		localname, _ := vid.Download()
		log.Println(localname)
	}
}

```

## [Library] go.mod

```
	github.com/heilkit/tt v1.1.0
```