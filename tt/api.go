package tt

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var (
	URL              string        = "https://tikwm.com/api"
	Timeout          time.Duration = time.Second + time.Millisecond*100
	MaxUserFeedCount int           = 33
	Debug                          = false
	requestSync      *sync.Mutex   = &sync.Mutex{}
)

func Raw(method string, query map[string]string) ([]byte, error) {
	if Timeout != 0 {
		requestSync.Lock()
		defer unlock()
	}

	url := fmt.Sprintf("%s/%s", URL, method)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	for key, val := range query {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buffer, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if Debug {
		log.Print(string(buffer))
	}

	return buffer, nil
}

func RawParsed[T any](method string, query map[string]string) (*T, error) {
	data, err := Raw(method, query)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code          int     `json:"code"`
		Msg           string  `json:"msg"`
		ProcessedTime float64 `json:"processed_time"`
		Data          *T      `json:"data"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		queryStr := "???"
		if buf, err := json.Marshal(query); err == nil {
			queryStr = string(buf)
		}
		return nil, fmt.Errorf("tikwm error: %s (%d) [%s, query: %s]", resp.Msg, resp.Code, method, queryStr)
	}

	return resp.Data, nil
}

// GetPost (hd default: true)
func GetPost(url string, hd ...bool) (*Post, error) {
	query := map[string]string{"url": url}
	if len(hd) == 0 || hd[0] {
		query["hd"] = "1"
	}
	return RawParsed[Post]("", query)
}

// GetUserFeedRaw is almost unuseful by itself, check wrappers around it -- GetUserFeed/GetUserFeedAwait.
func GetUserFeedRaw(uniqueID string, count int, cursor string) (*UserFeed, error) {
	query := map[string]string{"unique_id": uniqueID, "count": strconv.Itoa(count), "cursor": cursor}
	if _, err := strconv.ParseInt(uniqueID, 10, 64); err == nil {
		query = map[string]string{"user_id": uniqueID, "count": strconv.Itoa(count), "cursor": cursor}
	}
	return RawParsed[UserFeed]("user/posts", query)
}

func GetUserDetail(uniqueID string) (*UserDetail, error) {
	query := map[string]string{"unique_id": uniqueID}
	return RawParsed[UserDetail]("user/info", query)
}

func unlock() {
	time.Sleep(Timeout)
	requestSync.Unlock()
}
