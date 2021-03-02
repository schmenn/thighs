package main

import (
	"bufio"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"strings"
)

const (
	OAuthConsumerKey = "OAUTH_CONSUMER_KEY"
	OAuthConsumerSecret = "OAUTH_CONSUMER_SECRET"
	OAuthToken = "OAUTH_TOKEN"
	OAuthTokenSecret = "OAUTH_TOKEN_SECRET"	
)

func end() {
	fmt.Println("\nProgram finished. Press ENTER to close this window...")
	_, _ = fmt.Scanln()
	os.Exit(0)
}

func waitForInput(msg string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(msg)
	text, _ := reader.ReadString('\n')
	text = strings.ReplaceAll(text, "\n", "")
	text = strings.ReplaceAll(text, "\r", "")
	return text
}

func checkResponse(str string, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(str), prefix)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Could not load .env file; it might be missing. Add it to your project root.")
		end()
		return
	}
	for _, v := range [4]string{
		OAuthConsumerKey,
		OAuthConsumerSecret,
		OAuthToken,
		OAuthTokenSecret,
	} {
		if os.Getenv(v) == "" {
			fmt.Println(v + " is missing from your environment configuration. Ensure it is set, then try again.")
			end()
			return
		}
	}
	config := oauth1.NewConfig(os.Getenv(OAuthConsumerKey), os.Getenv(OAuthConsumerSecret))
	token := oauth1.NewToken(os.Getenv(OAuthToken), os.Getenv(OAuthTokenSecret))
	httpClient := config.Client(oauth1.NoContext, token)
	client := twitter.NewClient(httpClient)

	fmt.Println("=========\nTwitter Custom Source Tool\n=========")

	isTweetReply := waitForInput("Are you replying to a tweet? [y/n] ")
	var replyTo int64
	var replyToUsername string
	if checkResponse(isTweetReply, "y") {
		getRequestedID := waitForInput("What is the ID of the tweet you are replying to? ")
		a, err := strconv.ParseInt(getRequestedID, 0, 64)
		if err != nil {
			fmt.Println("Not a valid tweet ID:", err)
			end()
			return
		}
		fmt.Println("Fetching tweet to make sure it exists...")
		t, _, err := client.Statuses.Show(a, nil)
		if err != nil {
			fmt.Println(err)
			end()
			return
		}
		if t == nil {
			fmt.Println("Tweet doesn't exist")
			end()
			return
		}
		fmt.Println("Done")
		replyTo = t.ID
		replyToUsername = t.User.ScreenName
	}

	tweetText := waitForInput("Enter text that you want to send, then press ENTER: ")
	wantsToSetLoc := waitForInput("Do you want to set a custom Tweet location (latitude/longitude)? [y/n] ")
	var lat *float64 = nil
	var long *float64 = nil
	if checkResponse(wantsToSetLoc, "y") {
		wantsPreMadeList := waitForInput("Do you want to select from a pre-made list of locations (n for custom location)? [y/n] ")
		if checkResponse(wantsPreMadeList, "y") {
			var placesStr []string
			for i, p := range Places {
				placesStr = append(placesStr, fmt.Sprintf("[%d] %s (lat: %f, long: %f)", i + 1, p.Name, p.Lat, p.Long))
			}
			fmt.Println(strings.Join(placesStr, "\n"))
			var resAsInt int
			fmt.Println("Type a number then press ENTER: ")
			_, e := fmt.Scanln(&resAsInt)
			if e != nil {
				fmt.Println("Could not parse number.")
				end()
				return
			}

			if resAsInt < 0 || resAsInt > len(Places) {
				fmt.Println("Out of bounds.")
				end()
				return
			}
			lat = &Places[resAsInt - 1].Lat
			long = &Places[resAsInt - 1].Long
		} else {
			args := waitForInput("Input desired latitude and longitude number, separated by a space (5.000 6.000): ")
			latLongArray := strings.Split(args, " ")
			if len(latLongArray) != 2 {
				fmt.Println("Latitude/longitude not correctly formatted.")
				end()
				return
			}
			latNum, err := strconv.ParseFloat(latLongArray[0], 64)
			if err != nil {
				fmt.Println("Failed to parse latitude.")
				end()
				return
			}
			longNum, err := strconv.ParseFloat(strings.TrimRight(latLongArray[1], "\n\r"), 64)
			if err != nil {
				fmt.Println("Failed to parse longitude.")
				end()
				return
			}
			lat = &latNum
			long = &longNum
		}
	}

	fmt.Println("Sending tweet...")
	if replyTo != 0 {
		tweetText = fmt.Sprintf("@%s %s", replyToUsername, tweetText)
	}

	tweet, _, err := client.Statuses.Update(tweetText, &twitter.StatusUpdateParams{
		Status:             "",
		InReplyToStatusID:  replyTo,
		PossiblySensitive:  nil,
		Lat:                lat,
		Long:               long,
		PlaceID:            "",
		TrimUser:           nil,
		MediaIds:           nil,
		TweetMode:          "",
	})
	if err != nil {
		fmt.Println("Failed to send tweet. " + err.Error())
		end()
		return
	}
	fmt.Println("Done! Access your tweet at:\n" + fmt.Sprintf("https://twitter.com/%s/status/%s", tweet.User.ScreenName, tweet.IDStr))
	end()
	return
}
