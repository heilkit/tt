package tt

import (
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"net/http"
	"os"
	"path"
	"sync"
	"time"
)

var (
	DefaultDownloadClient = &grab.Client{
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			},
			Timeout: time.Minute * 15,
		},
		// UserAgent from https://explore.whatismybrowser.com/useragents/parse/505617920-tiktok-android-webkit,
		// UserAgent: "Mozilla/5.0 (Linux; Android 13; 2109119DG Build/TKQ1.220829.002; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/119.0.6045.193 Mobile Safari/537.36 trill_320403 JsSdk/1.0 NetType/WIFI Channel/googleplay AppName/trill app_version/32.4.3 ByteLocale/en ByteFullLocale/en Region/MY AppId/1180 Spark/1.4.6.3-bugfix AppVersion/32.4.3 BytedanceWebview/d8a21c6",
		// This UserAgent should be less trackable
		UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.6 Safari/605.1.1",
	}
	DefaultDownloadTimeout        = time.Millisecond * 100
	DefaultDownloadTimeoutOnError = time.Second * 30
	DefaultRetries                = 2
	FilenameFormat                = formatFilename
	downloadSync                  = &sync.Mutex{}
)

func Download(url string, opts ...DownloadOpt) (post *Post, files []string, err error) {
	opt := &DownloadOpt{}
	if len(opts) != 0 {
		opt = &opts[0]
	}
	opt = opt.Defaults()

	post, err = GetPost(url, true)
	if err != nil {
		return nil, nil, err
	}

	files, err = post.Download(opts...)
	return post, files, err
}

func (post Post) DownloadVideo(opts ...DownloadOpt) (file string, err error) {
	posts, err := post.Download(opts...)
	if len(posts) == 0 {
		return "", err
	}
	return posts[0], err
}

func (post Post) Download(opts ...DownloadOpt) (files []string, err error) {
	opt := &DownloadOpt{}
	if len(opts) != 0 {
		opt = &opts[0]
	}
	opt = opt.Defaults()

	if !opt.NoSync {
		downloadSync.Lock()
		defer downloadSync.Unlock()
	}

	urls := post.ContentUrls()

	files = []string{}
	for i, _ := range urls {
		to := opt.To
		if to == "" {
			tmp, err := os.Create(path.Join(opt.Directory, FilenameFormat(&post, i)))
			if err != nil {
				return files, err
			}
			files = append(files, tmp.Name())
			if err := tmp.Close(); err != nil {
				return files, err
			}
			to = tmp.Name()
		}

		if i > 0 {
			time.Sleep(opt.Timeout)
		}
		if err := opt.downloadRetrying(&post, i, to, 0, nil); err != nil {
			for _, file := range files {
				_ = os.Remove(file)
			}
			opt.NoSync = true
			return opt.Fallback(&post, *opt, err)
		}
	}

	return
}

func formatFilename(post *Post, i int) string {
	filename := fmt.Sprintf("%s_%s_%s", post.Author.UniqueId, time.Unix(post.CreateTime, 0).Format(time.DateOnly), post.ID())
	if post.IsVideo() {
		return filename + ".mp4"
	}
	return fmt.Sprintf("%s_%d.jpg", filename, i+1)
}

func (opt *DownloadOpt) downloadRetrying(post *Post, i int, filename string, try int, lastErr error) error {
	if try > opt.Retries {
		return lastErr
	}

	url := post.ContentUrls()[i]
	ret := func(err error) error {
		if try != opt.Retries {
			time.Sleep(opt.TimeoutOnError)
		}
		retry, retryErr := GetPost(post.ID())
		if retryErr != nil {
			return opt.downloadRetrying(retry, i, filename, try+1, retryErr)
		}
		*post = *retry
		return opt.downloadRetrying(retry, i, filename, try+1, err)
	}

	if err := opt.DownloadWith(url, filename); err != nil {
		return ret(err)
	}

	if valid, err := opt.ValidateWith(filename); err != nil {
		return ret(err)
	} else if !valid {
		return ret(err)
	}

	return nil
}
