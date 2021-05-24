package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"twitter-sender/utils"
)

const (
	OAuthConsumerKey     = "OAUTH_CONSUMER_KEY"
	OAuthConsumerSecret  = "OAUTH_CONSUMER_SECRET"
	OAuthToken           = "OAUTH_TOKEN"
	OAuthTokenSecret     = "OAUTH_TOKEN_SECRET"
	TwitterMediaAPI      = "https://upload.twitter.com"
	TwitterMaxCharacters = 280
	StatusCheckCap       = 10
	NewChunkWaitTime     = 25
	PlacesFile           = "places.json"
	TweetImage           = "tweet_image"
	TweetVideo           = "tweet_video"
	TweetGif             = "tweet_gif"
	Version              = "0.0.2"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		color.HiRed("[!] Could not load .env file; it might be missing. Add it to your project root.")
		return
	}
	for _, v := range [4]string{
		OAuthConsumerKey,
		OAuthConsumerSecret,
		OAuthToken,
		OAuthTokenSecret,
	} {
		if os.Getenv(v) == "" {
			color.HiRed("[!] %s is missing from your environment configuration. Ensure it is set, then try again.", v)
			return
		}
	}

	config := oauth1.NewConfig(os.Getenv(OAuthConsumerKey), os.Getenv(OAuthConsumerSecret))
	token := oauth1.NewToken(os.Getenv(OAuthToken), os.Getenv(OAuthTokenSecret))
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	latPtr := flag.Float64("lat", 0.0, "The latitude of the tweet")
	longPtr := flag.Float64("long", 0.0, "The longitude of the tweet")
	replyToPtr := flag.Int64("replyto", 0, "If you are replying to a tweet, specify its ID here")
	mediaPtr := flag.String("media", "", "A list of media files following the format 1,2,3,4. Needs to be an image, video, or GIF. Note that you can either: use 4 images OR 1 video OR 1 GIF per tweet. ")
	debug := flag.Bool("debug", false, "Specify this to view detailed logs")
	placeIDPtr := flag.String("placeid", "", "If you have places.json in your folder, specify the ID to use.")

	flag.Parse()

	if *debug {
		color.Blue("debug mode -- more logs will be shown")
	}

	tweetText := flag.Args()
	tweetTextJoined := strings.Join(tweetText, " ")

	color.HiBlue("---- thighs: Custom Twitter source messages version %s ----\nhttps://github.com/vysiondev/thighs", Version)

	if *debug {
		color.HiBlack("[debug] tweet text: %s", tweetTextJoined)
	}

	places, err := ParsePlacesFile()
	if err != nil {
		color.HiRed("Could not parse your places.json file. %s", err.Error())
		return
	}

	if *placeIDPtr != "" {
		if places == nil {
			color.HiYellow("[!] You specified -placeid, but do not have a places.json file that was parsed. It will be ignored.")
		} else {
			match := false
			for _, p := range *places {
				if strings.ToLower(p.ID) == strings.ToLower(*placeIDPtr) {
					*latPtr = p.Lat
					*longPtr = p.Long
					match = true
					color.White("Using location %s with lat: %f, long: %f", p.Name, p.Lat, p.Long)
					break
				}
			}
			if !match {
				color.HiYellow("[!] You specified -placeid, but none of your places matched the ID you specified. It will be ignored.")
			}
		}
	}

	var replyToUsername string
	if *replyToPtr != 0 {
		if *debug {
			color.HiBlack("[debug] user will reply to tweet with id %d", *replyToPtr)
			color.HiBlack("[debug] fetching tweet to make sure it exists...")
		}
		t, _, err := client.Statuses.Show(*replyToPtr, nil)
		if err != nil {
			color.HiRed("[!] Encountered an error while trying to fetch tweet to reply to: %s", err.Error())
			return
		}
		if t == nil {
			color.HiRed("[!] Reply tweet ID doesn't exist or it's not viewable to the public.")
			return
		}
		if *debug {
			color.HiBlack("[debug] tweet found; setting reply to mention @%s", t.User.ScreenName)
		}
		*replyToPtr = t.ID
		replyToUsername = t.User.ScreenName
	}
	if *debug {
		color.HiBlack("[debug] completed reply check")
	}

	if *replyToPtr != 0 {
		tweetTextJoined = fmt.Sprintf("@%s %s", replyToUsername, tweetTextJoined)
	}

	if len(tweetText) >= TwitterMaxCharacters {
		color.HiRed("[!] Your proposed Tweet would be too long (%d characters >= 280). Try making it shorter.", len(tweetText))
		return
	}
	if *debug {
		color.HiBlack("[debug] completed message length check")
	}

	mediaIds := make([]int64, 0)
	if len(*mediaPtr) > 0 {
		specialFileUploaded := false
		mediaSplit := strings.Split(*mediaPtr, ",")
		if len(mediaSplit) == 0 {
			color.HiRed("[!] You didn't specify any media to upload.")
			return
		}
		if *debug {
			color.HiBlack("[debug] %d media files to upload", len(mediaSplit))
		}
		for _, f := range mediaSplit {
			if specialFileUploaded == true {
				color.HiYellow("[!] An image/video was already processed; skipping all other files.")
				break
			}
			color.White("Uploading %s", f)
			file, err := os.Open(f)
			if err != nil {
				color.HiRed("[!] Failed to open file %s: %s.", f, err.Error())
				return
			}
			defer file.Close()
			fstat, err := file.Stat()
			if err != nil {
				color.HiRed("[!] Failed to stat file %s: %s.", f, err.Error())
				return
			}
			if fstat.Size() >= 15*1024*1024 {
				color.HiRed("[!] File is too big (>=15MB).")
				return
			}
			// get file content type
			contentType, err := utils.GetFileContentType(file)
			if err != nil {
				color.HiYellow("[!] Failed to detect content type for %s: %s. Will use application/octet-stream instead.", f, err.Error())
				contentType = "application/octet-stream"
			}
			if *debug {
				color.HiBlack("[debug] this file is of type %s", contentType)
				color.HiBlack("[debug] target size reported by stat: %d (all chunks should add up to this number)", fstat.Size())
				color.HiBlack("[debug] resetting reader to 0")
			}
			_, err = file.Seek(0, io.SeekStart)
			if err != nil {
				color.HiRed("[!] Failed to seek file reader to 0: %s.", err.Error())
				return
			}
			if *debug {
				color.HiBlack("[debug] making INIT request")
			}
			mediaID, err := CallInit(httpClient, fstat.Size(), contentType)
			if err != nil {
				color.HiRed("[!] Failed on INIT request: %s.", err.Error())
				return
			}
			if *debug {
				color.HiBlack("[debug] INIT successful; got media ID %d", mediaID)
			}
			var segmentID int64
			reader := bufio.NewReader(file)

			chunkSize := utils.CalculateChunkSize(fstat.Size())
			if *debug {
				color.HiBlack("[debug] chunk size (%d * 0.30) is %d", fstat.Size(), chunkSize)
			}
			buf := make([]byte, 0, chunkSize)
			needToWait := false
			appendResponseChan := make(chan *AppendResponse)
			var wg sync.WaitGroup

			if *debug {
				color.HiBlack("[debug] upload starting")
			}
			for {
				n, err := io.ReadFull(reader, buf[:cap(buf)])
				buf = buf[:n]
				if *debug {
					color.HiBlack("[debug] [async:%d] %d bytes to be uploaded", segmentID, len(buf))
				}
				if err != nil {
					if err != io.EOF && err != io.ErrUnexpectedEOF {
						color.HiRed("[!] Failed to read the file into a buffer because of something other than EOF: %s.", err.Error())
						return
					}
					wg.Add(1)
					go CallAppend(mediaID, segmentID, &buf, httpClient, appendResponseChan, &wg)
					break
				}

				wg.Add(1)
				go CallAppend(mediaID, segmentID, &buf, httpClient, appendResponseChan, &wg)
				segmentID++

				// not sure what's causing chunks to write the wrong # of bytes unless there's a short delay...
				time.Sleep(time.Millisecond * time.Duration(NewChunkWaitTime))
			}

			go func() {
				wg.Wait()
				close(appendResponseChan)
			}()

			for a := range appendResponseChan {
				if a.StatusCode == 0 {
					color.HiRed("[!] Async upload has failed.")
					return
				}
				if a.StatusCode < 200 || a.StatusCode > 299 {
					color.HiRed("[!] Async upload hit a response that reported status code %s (not 2xx).", a.StatusCode)
					return
				}
				if *debug {
					color.HiBlack("[debug] [async:%d] done; status %d", a.SegmentID, a.StatusCode)
				}
			}

			if *debug {
				color.HiBlack("[debug] upload finished")
			}

			wait, e := CallFinalize(mediaID, httpClient)
			if e != nil {
				color.HiRed("[!] Error while calling FINALIZE: %s.", e.Error())
				return
			}
			needToWait = wait
			if *debug {
				color.HiBlack("[debug] FINALIZE call successful")
				if needToWait {
					color.HiBlack("[debug] FINALIZE says upload is not done processing; we need to wait")
				}
			}

			if needToWait {
				statusChecks := 0
				var status *MediaStatusResponse
				for {
					if statusChecks > StatusCheckCap {
						color.HiRed("[!] Waited for too long for upload to complete. Bailing out")
						return
					}
					if *debug {
						color.HiBlack("[debug] calling STATUS on media id %d (try %d of %d)", mediaID, statusChecks+1, StatusCheckCap)
					}
					statusObject, err := CallStatus(mediaID, httpClient)
					if err != nil {
						color.HiRed("[!] Error while checking for status: %s.", err.Error())
						return
					}
					status = statusObject
					if statusObject.ProcessingInfo.State != "in_progress" {
						break
					}
					color.HiBlack("File is being processed (progress: %d%%). Checking again after %d seconds", statusObject.ProcessingInfo.ProgressPercent, statusObject.ProcessingInfo.CheckAfterSecs)
					statusChecks++
					time.Sleep(time.Second * time.Duration(status.ProcessingInfo.CheckAfterSecs))
				}
				if status == nil {
					color.HiRed("[!] No data on status object. This could indicate that it failed to upload.")
					return
				}
				if status.ProcessingInfo.State != "succeeded" {
					color.HiRed("[!] Upload of %s was not successful (status: %s): %s: %s", f, status.ProcessingInfo.State, status.ProcessingInfo.Error.Name, status.ProcessingInfo.Error.Message)
					return
				}

			}

			if strings.Contains(contentType, "video") || strings.Contains(contentType, "gif") {
				if *debug {
					color.HiBlack("[debug] special media type, so this is the only one that will be used in the tweet (overriding everything else)")
				}
				mediaIds = make([]int64, 0)
				specialFileUploaded = true
			}
			color.HiGreen("%s uploaded & processed successfully", f)
			mediaIds = append(mediaIds, mediaID)
		}
	}
	if *debug {
		color.HiBlack("[debug] completed upload of all files")
	}

	color.White("Sending tweet...")
	tweet, _, err := client.Statuses.Update(tweetTextJoined, &twitter.StatusUpdateParams{
		Status:            "",
		InReplyToStatusID: *replyToPtr,
		PossiblySensitive: nil,
		Lat:               latPtr,
		Long:              longPtr,
		PlaceID:           "",
		TrimUser:          nil,
		MediaIds:          mediaIds,
		TweetMode:         "",
	})
	if err != nil {
		color.HiRed("[!] Failed to send tweet. %s", err.Error())
		return
	}
	color.HiGreen("Tweet successfully posted! You can find it at:\n" + fmt.Sprintf("https://twitter.com/%s/status/%s", tweet.User.ScreenName, tweet.IDStr))
	return
}
