package tt

import (
	"log"
	"time"
)

type Predicate func(post *Post) bool

func FilterVideo(post *Post) bool {
	return post.IsVideo()
}

func FilterPhoto(post *Post) bool {
	return post.IsAlbum()
}

var _ Predicate = FilterVideo
var _ Predicate = FilterPhoto

func WhileAfter(t time.Time) Predicate {
	return func(post *Post) bool {
		createTime := time.Unix(post.CreateTime, 0)
		return createTime.After(t)
	}
}

type FeedOpt struct {
	// Filter -- classic filter function (default: add all)
	Filter Predicate
	// While to continue scanning (default: scan all)
	While Predicate
	// OnError could panic to interrupt the job (default: log the error)
	OnError func(err error)
	// ReturnChan == nil, then it will be created inside the function.
	// ReturnChan is closed when scanning subroutine is done.
	ReturnChan chan Post
	SD         bool
}

func (opt *FeedOpt) Defaults() *FeedOpt {
	if opt == nil {
		opt = &FeedOpt{}
	}
	if opt.Filter == nil {
		opt.Filter = func(vid *Post) bool { return true }
	}
	if opt.While == nil {
		opt.While = func(vid *Post) bool { return true }
	}
	if opt.OnError == nil {
		opt.OnError = func(err error) {
			log.Print(err)
		}
	}
	if opt.ReturnChan == nil {
		opt.ReturnChan = make(chan Post)
	}
	return opt
}

func GetUserFeedAwait(uniqueID string, opts ...FeedOpt) ([]Post, error) {
	postChan, _, err := GetUserFeed(uniqueID, opts...)
	if err != nil {
		return nil, err
	}
	ret := []Post{}
	for post := range postChan {
		ret = append(ret, post)
	}
	return ret, nil
}

// GetUserFeed to a channel, getting HD versions of files could take a while.
// If you are okay with waiting for minutes ((1-2 secs) * len_of_videos), consider GetUserFeedAwait
func GetUserFeed(uniqueID string, opts ...FeedOpt) (chan Post, int, error) {
	var opt *FeedOpt = nil
	if len(opts) != 0 {
		opt = &opts[0]
	}
	opt = opt.Defaults()

	posts, err := userFeedUntilInternal(uniqueID, "0", opt)
	if err != nil {
		return nil, 0, err
	}
	for i := 0; i < len(posts)/2; i++ {
		posts[i], posts[len(posts)-i-1] = posts[len(posts)-i-1], posts[i]
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// It's ok to panic in onError to interrupt the loop,
				// nothing else would panic (95% sure).
			}
		}()

		defer close(opt.ReturnChan)
		for _, post := range posts {
			if opt.SD {
				opt.ReturnChan <- post
				continue
			}

			vidHD, err := GetPost(post.VideoId, true)
			if err != nil {
				opt.OnError(err)
				opt.ReturnChan <- post
			} else {
				opt.ReturnChan <- *vidHD
			}
		}
	}()

	return opt.ReturnChan, len(posts), err
}

func userFeedUntilInternal(uniqueID string, cursor string, opt *FeedOpt) ([]Post, error) {
	feed, err := GetUserFeedRaw(uniqueID, MaxUserFeedCount, cursor)
	if err != nil {
		return nil, err
	}

	ret := []Post{}
	if len(feed.Videos) > MaxUserFeedCount {
		feed.Videos = feed.Videos[:MaxUserFeedCount]
	}
	for _, vid := range feed.Videos {
		if !opt.While(&vid) {
			return ret, nil
		}
		if opt.Filter(&vid) {
			ret = append(ret, vid)
		}
	}

	if !feed.HasMore {
		return ret, nil
	}

	deeperRet, err := userFeedUntilInternal(uniqueID, feed.Cursor, opt)
	if err != nil {
		return ret, err
	}

	return append(ret, deeperRet...), nil
}
