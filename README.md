# thighs

Send Tweets from custom sources! You can attach media and reply to others as well. Utilizes async uploading for faster completion times. Also has a really good name.

![](https://i.imgur.com/bmG2fsl.png)

## Setup

1. Download the binary for your OS or compile it from source.   
2. Get your consumer key & secret from the Twitter developer portal.
3. Get an access token & secret for yourself so that the tool can send a Tweet on your behalf.
4. Create a .env file in the project root and fill these values in:

```
# API key
OAUTH_CONSUMER_KEY=
# API secret
OAUTH_CONSUMER_SECRET=
# Access token
OAUTH_TOKEN=
# Token secret
OAUTH_TOKEN_SECRET=
```

## Reference

|Option|Description|
|-----|-----|
|`-replyto`|If you're replying to another Tweet, use its ID here|
|`-media`|A list of media file paths to use. They must be separated with commas (`1,2,3...`). Wrap the argument in double quotes if there are spaces in the file names. If you specify this, you don't need to put in any tweet text. 4 images can be uploaded per tweet **OR** a single video **OR** a single GIF. **The first GIF/video in the media list will be used in case more media is specified after it.**|
|`-lat`|Latitude of the tweet location|
|`-long`|Longitude of the tweet location|
|`-placeid`|**see the next section for more information on places.json**. If you have `places.json` in your folder, specify the index of a place to use, referencing the `id` field of the place. For example, if you set a place id to `cool`, use `-placeid cool`. This will override `-lat` and `-long` if they were manually specified.|
|`-debug`|Shows more logs. Does not need any additional arguments.|
|`-chunk-interval`|Change the interval (in ms) at which a buffer is read from media and sent as an asynchronous POST request. By default, it is 25 ms. Try changing this to a higher number if your uploads are failing.

## Places.json

You can use `places.json` to quickly access a pair of coordinates based on an ID that you give it.

1. Download `places.json` to the same folder as the binary.
2. Edit it if you want. It must follow the same format as the places already in the JSON file.
3. Run the program with `-placeid`. See the reference section above. 

### Examples

Reply to a Tweet with message "sussy":

`thighs -replyto 1396269995630350340 sussy`

Add 2 images to a Tweet with message "sus af", and set the lat/long of the post to 100:

`thighs -media "image 1.png,image 2.png" -lat 100.0 -long 100.0 sus af`

### Notes

- If you want to put in mentions, just @mention the person like normal.
- Quotes can be typed in tweet text using `\"` and `\'`.

