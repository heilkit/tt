# Tikwm API

https://tikwm.com is the best middleman for getting TikTok video info, afaik. If I'm wrong, contact me immediately.

* Download videos in HD
* Download whole profiles in a minimal time (with reasonable naming)

## [Library] Example

```go
package main

import (
  "github.com/heilkit/tt/tt"
  "log"
  "time"
)

func main() {
  // tt.GetVideo(url string, HD bool) ()
  postHD, err := tt.GetPost("https://www.tiktok.com/@locallygrownwig/video/6901498776523951365")
  postHD, err = tt.GetPost("6901498776523951365", true)                // with ID 
  postSD, err := tt.GetPost("https://vm.tiktok.com/ZM66UoB9m/", false) // with shorten link 
  localname, err := postHD.DownloadVideo(tt.DownloadOpt{To: "locallygrownwig.mp4"})

  // Get user posts for the last 30 days
  until := time.Now().Add(-time.Hour * 24 * 30)
  vidChan, expectedCount, err := tt.GetUserFeed("locallygrownwig", &tt.FeedOpt{
    While:  tt.WhileAfter(until),
    Filter: tt.FilterVideo,
  })

  for vid := range vidChan {
    localname, _ := vid.DownloadVideo()
    log.Println(localname)
  }
}

```

## [Executable] Example

* `./tikmeh "https://www.tiktok.com/@locallygrownwig/video/6901498776523951365"` -- download this video in HD to current
  folder
* `./tikmeh -profile -until "2023-01-01 00:00:00" losertron` -- download all @losertron content from 2023 to now
* `./tikmeh -info losertron` -- get user info about @losertron profile

```
$ ./tikmeh
Usage: ./tikmeh [-profile | -info] [args...] <urls | usernames | ids>
  -info
        print info about profiles
  -profile
        download/scan profiles
  -dir string
        directory to save files (default "./")
  -debug
        log debug info
  -json
        print info as json, don't download
  -quiet
        quiet
  -sd
        don't request HD sources of videos (less requests => notably faster)
  -until string
        don't download videos earlier than (default "1970-01-01 00:00:00")
```
