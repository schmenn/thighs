package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"twitter-sender/utils"
)

func CallInit(c *http.Client, size int64, contentType string) (int64, error) {
	var mediaCategory string
	switch contentType {
	case "image/png", "image/jpg", "image/jpeg", "image/webp":
		mediaCategory = TweetImage
		break
	case "image/gif":
		mediaCategory = TweetGif
	}
	if strings.Contains(contentType, "video") {
		mediaCategory = TweetVideo
	}
	res, err := c.PostForm(TwitterMediaAPI+"/1.1/media/upload.json", url.Values{
		"command":        {"INIT"},
		"total_bytes":    {strconv.FormatInt(size, 10)},
		"content_type":   {contentType},
		"media_category": {mediaCategory},
	})
	if err != nil {
		return 0, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		_ = res.Body.Close()
		return 0, errors.New("non 2xx response code: " + strconv.FormatInt(int64(res.StatusCode), 10))
	}
	var b *MediaInitResponse
	err = json.NewDecoder(res.Body).Decode(&b)
	if err != nil {
		_ = res.Body.Close()
		return 0, err
	}
	_ = res.Body.Close()
	return b.MediaId, nil
}

type AppendResponse struct {
	SegmentID  int64
	StatusCode int
}

func CallAppend(mediaID int64, segmentID int64, buf *[]byte, c *http.Client, ch chan *AppendResponse, wg *sync.WaitGroup) {
	defer (*wg).Done()
	req, err := utils.NewMultipartForm(TwitterMediaAPI+"/1.1/media/upload.json", map[string]string{
		"command":       "APPEND",
		"media_id":      strconv.FormatInt(mediaID, 10),
		"segment_index": strconv.FormatInt(segmentID, 10),
	}, buf)
	if err != nil {
		ch <- &AppendResponse{
			SegmentID:  segmentID,
			StatusCode: 0,
		}
		return
	}
	res, err := c.Do(req)
	if err != nil {
		ch <- &AppendResponse{
			SegmentID:  segmentID,
			StatusCode: 0,
		}
		return
	}
	_ = res.Body.Close()
	ch <- &AppendResponse{
		SegmentID:  segmentID,
		StatusCode: res.StatusCode,
	}
	return
}

func CallFinalize(mediaID int64, c *http.Client) (bool, error) {
	res, err := c.PostForm(TwitterMediaAPI+"/1.1/media/upload.json", url.Values{
		"command":  {"FINALIZE"},
		"media_id": {strconv.FormatInt(mediaID, 10)},
	})
	if err != nil {
		return false, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		_ = res.Body.Close()
		return false, errors.New("non 2xx response code: " + strconv.FormatInt(int64(res.StatusCode), 10))
	}
	var b *MediaFinalizeResponse
	err = json.NewDecoder(res.Body).Decode(&b)
	if err != nil {
		_ = res.Body.Close()
		return false, err
	}
	_ = res.Body.Close()
	if b.ProcessingInfo != nil {
		return true, nil
	}
	return false, nil
}

func CallStatus(mediaID int64, c *http.Client) (*MediaStatusResponse, error) {
	statusURL, _ := url.Parse(TwitterMediaAPI + "/1.1/media/upload.json")
	params := url.Values{}
	params.Add("command", "STATUS")
	params.Add("media_id", strconv.FormatInt(mediaID, 10))
	statusURL.RawQuery = params.Encode()

	res, err := c.Get(statusURL.String())
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		_ = res.Body.Close()
		return nil, errors.New("non 2xx response code: " + strconv.FormatInt(int64(res.StatusCode), 10))
	}
	var status *MediaStatusResponse
	err = json.NewDecoder(res.Body).Decode(&status)
	if err != nil {
		_ = res.Body.Close()
		return nil, err
	}

	_ = res.Body.Close()
	return status, nil
}
