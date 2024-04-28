package tt

import (
	"encoding/json"
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"path"
	"sync"
	"time"
)

var (
	DefaultDownloadMutex          = &sync.Mutex{}
	DefaultDownloadTimeout        = time.Second
	DefaultDownloadTimeoutOnError = time.Second * 15
	DefaultDownloadGrabClient     = defaultGrab()
)

// DownloadOpt provides verbose options for how you want to download posts.
type DownloadOpt struct {
	// Filename lets you set fixed file name, otherwise use FilenameFormat.
	Filename string
	// Directory download to.
	Directory string
	// DownloadWith specifies how singe files are downloaded.
	DownloadWith func(url string, filename string) error
	// ValidateWith function your downloads, by default do nothing.
	ValidateWith func(filename string) (bool, error)
	// Fallback in case something goes wrong, by default there's no Fallback. tt.FallbackToSD from the package.
	Fallback func(post *Post, opt DownloadOpt, err error) (files []string, e error)
	// FilenameFormat defaults to i.e. "canthinky_2022-12-21_7179438804418268417.mp4", unless other function is provided.
	FilenameFormat func(post *Post, i int) string
	// Timeout set to 0 is actually gets set to time.Second, if you want real no timeout use Timeout=time.Nanosecond.
	Timeout time.Duration
	// TimeoutOnError set to 0 is actually gets set to time.Second, if you want real no timeout use Timeout=time.Nanosecond.
	TimeoutOnError time.Duration
	// NoSync allows parallel downloads, by default it's disallowed.
	NoSync bool
	// Retries set to 0 is interpreted as no value, so it's getting set to DownloadDefaultReties. To have to Retries set it to -1.
	Retries int
	// Download post in SD quality.
	SD bool
	// Log if you need it.
	Log *slog.Logger
}

func (opt *DownloadOpt) WithDefaults() *DownloadOpt {
	if opt == nil {
		opt = &DownloadOpt{}
	}
	if opt.Directory == "" {
		opt.Directory = "."
	}
	if opt.Timeout == 0 {
		opt.Timeout = DefaultDownloadTimeout
	}
	if opt.TimeoutOnError == 0 {
		opt.TimeoutOnError = DefaultDownloadTimeoutOnError
	}
	if opt.Retries < 0 {
		opt.Retries = 0
	}
	if opt.Fallback == nil {
		opt.Fallback = fallbackNone
	}
	if opt.FilenameFormat == nil {
		if opt.Filename != "" {
			opt.FilenameFormat = DownloadTo(opt.Filename)
		} else {
			opt.FilenameFormat = formatFilename
		}
	}
	if opt.Log == nil {
		opt.Log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	if opt.DownloadWith == nil {
		opt.DownloadWith = DownloadFileWith
	}
	if opt.ValidateWith == nil {
		opt.ValidateWith = func(filename string) (bool, error) { return true, nil }
	}

	return opt
}

func Download(filename string, opt ...*DownloadOpt) (post *Post, filenames []string, err error) {
	opts := &DownloadOpt{}
	if len(opt) > 0 {
		opts = opt[0]
	}
	post, err = GetPost(filename, opts.SD)
	if err != nil {
		return nil, nil, fmt.Errorf("Download -> GetPost: %w", err)
	}
	files, err := post.Download(opts)
	return post, files, err
}

func DownloadSingle(filename string, opt ...*DownloadOpt) (filenames string, err error) {
	opts := &DownloadOpt{}
	if len(opt) > 0 {
		opts = opt[0]
	}
	post, err := GetPost(filename, opts.SD)
	if err != nil {
		return "", fmt.Errorf("DownloadSingle -> GetPost: %w", err)
	}
	files, err := post.Download(opts)
	return files[0], err
}

func (post Post) Download(opt ...*DownloadOpt) (filenames []string, err error) {
	opts := &DownloadOpt{}
	if len(opt) != 0 {
		opts = opt[0]
	}
	opts = opts.WithDefaults()
	if !opts.NoSync {
		DefaultDownloadMutex.Lock()
		defer DefaultDownloadMutex.Unlock()
	}

	for i, url := range post.ContentUrls(opts.SD) {
		time.Sleep(opts.Timeout)
		filename := path.Join(opts.Directory, opts.FilenameFormat(&post, i))
		if err := opts.DownloadWith(url, filename); err != nil {
			for try := 0; try < opts.Retries || err == nil; try++ {
				opts.Log.Warn("Download failed, retrying...", "err", err, "try", try+1)
				time.Sleep(opts.TimeoutOnError)
				err = opts.DownloadWith(url, filename)
			}
			//goland:noinspection GoDfaConstantCondition -- this is correct, bc `for` loop before ends if err == nil.
			if err != nil {
				return opts.Fallback(&post, *opts, fmt.Errorf("Download: %w", err))
			}
		}
		filenames = append(filenames, filename)
	}

	return filenames, err
}

func DownloadFileWith(url string, filename string) error {
	req, err := grab.NewRequest(filename, url)
	if err != nil {
		return fmt.Errorf("grab.NewRequest: %w", err)
	}

	if resp := DefaultDownloadGrabClient.Do(req); resp.Err() != nil {
		return fmt.Errorf("grab.Do: %w", resp.Err())
	}
	return nil
}

func DownloadTo(filename string) func(post *Post, i int) string {
	return func(post *Post, i int) string {
		return filename
	}
}

func FallbackToSD(post *Post, opt DownloadOpt, err error) (filenames []string, e error) {
	opt.Log.Warn("Downloading failed, falling back to SD", "post", post.ID(), "err", err)
	opt.SD = true
	opt.Fallback = fallbackNone
	return post.Download(&opt)
}

func fallbackNone(post *Post, opt DownloadOpt, err error) (files []string, e error) {
	return nil, err
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

func defaultGrab() *grab.Client {
	return &grab.Client{
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

}

func formatFilename(post *Post, i int) string {
	filename := fmt.Sprintf("%s_%s_%s", post.Author.UniqueId, time.Unix(post.CreateTime, 0).Format(time.DateOnly), post.ID())
	if post.IsVideo() {
		return filename + ".mp4"
	}
	return fmt.Sprintf("%s_%d.jpg", filename, i+1)
}
