# Twitter Custom Source Tool

Send Tweets from custom sources!

![](https://i.postimg.cc/yYSb0jkq/image.png)

Also, you can set your latitude/longitude to change the location. (You need to enable twitter to post your location data for this)

![](https://i.postimg.cc/gkDGJ1Q0/image.png)

### Setup

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

5. Double click the exe.