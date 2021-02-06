# Description

This bot tracks everyones presences, when they change the song they're listening to on spotify, it will add the data of that listening presnece to a database.

This allows for us to be able to get a leaderboard of what that user listens to.

# Setup

```
$ git clone https://github.com/postrequest69/spotify-presence-tracker.git
$ cd spotify-presence-tracker
$ go get github.com/bwmarrin/discordgo
Fillout config.go with your discord bot token and prefered prefix.
$ go run .
use the help command in your server :)
```

# TODO

- [ ] Add a command to get the leaderboard of another user, using user id.
- [ ] Change the database storage to sql from json
