package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	db      Database
	mainMap = make(map[string]*MainData)
)

// When the bot first launches, this event starts saving the database every 60 seconds and starts the 5 minute leaderboard page listener.
func ready(session *discordgo.Session, event *discordgo.Ready) {
	// update status
	session.UpdateGameStatus(0, "Just stalking spotify plays... don't mind me...")
	fmt.Println("Logged in as: " + event.User.Username + "#" + event.User.Discriminator)
	fmt.Println("Initilizing Database!")
	// read db
	input, err := ioutil.ReadFile("./database.json")
	if err != nil {
		fmt.Println("Error lol")
		return
	}
	// now we can use the db
	err = json.Unmarshal(input, &db)
	if err != nil {
		fmt.Println("xd nice error")
		return
	}
	// now we're converting it to a map to be used in the main functions
	mainMap = ConverToMap(db)
	// every 60 seconds we save the database
	go saveEvery60()
	// we check for no longer in use page listeners
	go checkListeners()
}

// removes page lsitenres that are no longer in use
func checkListeners() {
	for {
		time.Sleep((60 * 5) * time.Second)
		var now = MakeTimestamp()
		for key, listener := range pageListeners {
			if (now - listener.LastUsed) > 300000 {
				delete(pageListeners, key)
			}
		}
	}
}

// save the database every 60 seconds
func saveEvery60() {
	for {
		time.Sleep(60 * time.Second)
		toSave := ConvertMapToJSONStruct(mainMap)
		toSaveString, _ := json.Marshal(toSave)
		UpdateDatabase(string(toSaveString))
	}
}

// here is where we track spotify listening statuses
func presenceListen(session *discordgo.Session, event *discordgo.PresenceUpdate) {
	// make sure that they actualy have a status
	if len(event.Activities) > 0 {
		// go through every activity in slice to find the spotify one (if it exists)
		for _, activity := range event.Activities {
			// we found spotify presence
			if activity.Name == "Spotify" {
				//If user isn't in the current map database
				if _, ok := mainMap[event.User.ID]; !ok {
					//Make a user struct in map database
					member, err := session.User(event.User.ID)
					if err != nil {
						fmt.Println(err)
					}
					mainMap[event.User.ID] = &MainData{
						Userid:         event.User.ID,
						User:           member.Username + "#" + member.Discriminator,
						ListeningArray: make(map[string]*Artist),
					}
					mainMap[event.User.ID].ListeningArray[activity.State] = &Artist{
						ArtistName:   activity.State,
						TimesListend: 1,
						Songs:        make(map[string]*Song),
					}
					mainMap[event.User.ID].ListeningArray[activity.State].Songs[activity.Details] = &Song{
						SongName:      activity.Details,
						TimesListened: 1,
					}
					//done making struct :D
				} else {
					//If the user is in the current map
					//if the artist isn't int he current lsitjeing map
					if _, ok := mainMap[event.User.ID].ListeningArray[activity.State]; !ok {
						mainMap[event.User.ID].ListeningArray[activity.State] = &Artist{
							ArtistName:   activity.State,
							TimesListend: 1,
							Songs:        make(map[string]*Song),
						}
						mainMap[event.User.ID].ListeningArray[activity.State].Songs[activity.Details] = &Song{
							SongName:      activity.Details,
							TimesListened: 1,
						}
					} else {
						//if the artist is in the current listening map
						mainMap[event.User.ID].ListeningArray[activity.State].TimesListend++
						//song doesn't exist in song map in listening map
						if _, ok := mainMap[event.User.ID].ListeningArray[activity.State].Songs[activity.Details]; !ok {
							mainMap[event.User.ID].ListeningArray[activity.State].Songs[activity.Details] = &Song{
								SongName:      activity.Details,
								TimesListened: 1,
							}
						} else {
							//song does exist in song map in lsitening map
							mainMap[event.User.ID].ListeningArray[activity.State].Songs[activity.Details].TimesListened++
						}
					}
				}
			}
		}
	}
}

var (
	// top command help embed
	topHelpEmbed = &discordgo.MessageEmbed{
		Title:       "Help for top command!",
		Description: "The top command has one required argument and another optional argument!\n\n**__Required option choices__**: artist, song\n\n**__Artist__**: sets the top showing mode to artist, if you just want to see your top artists just do `" + Prefix + "top artist`\nIf you want the top songs from an artist name you can do `" + Prefix + "top artist artist name` (case sensitive)\n\nSong: if you just do `" + Prefix + "top song` you will be shown all your most listened songs. There is no optional argument for this.",
	}
	// the page listeners
	pageListeners = make(map[string]*ReactionListener)
)

const (
	// all the emojis we use
	leftArrow, rightArrow, destroyEmoji = "⬅️", "➡️", "❌"
)

// here is where we listen for reactions (for the pages)
func reactionListen(session *discordgo.Session, reaction *discordgo.MessageReactionAdd) {

	// if the message being reacted to is in the reaction map
	if _, ok := pageListeners[reaction.MessageID]; ok {
		// validating that the user reacting is indeed the user that owns the listener
		if pageListeners[reaction.MessageID].UserID != reaction.UserID {
			return
		}
		// switch the emoji name, name is misleading, basically -> ❌
		switch reaction.Emoji.Name {
		// when the reaction used is a left arrow (page decrease)
		case leftArrow:
			// update last used so the lsitener isn't deemed inactive
			pageListeners[reaction.MessageID].LastUsed = MakeTimestamp()
			// remove reaction, better user expirence
			session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, leftArrow, reaction.UserID)
			// If page is already 1 aka lowest it can be
			if pageListeners[reaction.MessageID].CurrentPage == 1 {
				break
			}
			// decrease current page
			pageListeners[reaction.MessageID].CurrentPage--
			// the max, we get this by getting the current page starting at index 1 and timsing it by 10 to set a limit
			var max = pageListeners[reaction.MessageID].CurrentPage * 10
			// we subtract one, this is because arrays and slices start at index 0, the max assums that it starts at index 1
			max--
			// the min is just the max minus 9, this collects all 10 between the scope
			var min = max - 9
			// if the leaderboard is for artists
			if pageListeners[reaction.MessageID].Type == "artist" {
				// this is for if a specific artist is listed, whole different output so we need to completely structre a method behind it
				if pageListeners[reaction.MessageID].Specific {
					// build a string to send in the embed, 1.) song lsitened x times\n2.) ...
					toSend := ConvertToSendAbleStringSpecific(pageListeners[reaction.MessageID].SpecificArtist, pageListeners[reaction.MessageID].SpecificArtistStruct, pageListeners[reaction.MessageID].Author, min, max)
					// we just edit the embed here so it updates
					session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
						Description: toSend,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "Page " + fmt.Sprint(pageListeners[reaction.MessageID].CurrentPage) + "/" + fmt.Sprint(pageListeners[reaction.MessageID].PageLimit),
						},
					})
				} else {
					// build a string to send in the embed, 1.) artist lsitened to x times\n2.) ...
					toSend, _ := ConvertToSendableString(pageListeners[reaction.MessageID].Data, "artist", min, max)
					// we just edit the embed here so it updates
					session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
						Description: toSend,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "Page " + fmt.Sprint(pageListeners[reaction.MessageID].CurrentPage) + "/" + fmt.Sprint(pageListeners[reaction.MessageID].PageLimit),
						},
					})
				}
			} else {
				// this is for the songs leaderboard, we build a string to send in the embed. 1.) song listened x times\n2.) ...
				toSend, _ := ConvertToSendableString(pageListeners[reaction.MessageID].Data, "song", min, max)
				// we just edit the embed here so it updates
				session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
					Description: toSend,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Page " + fmt.Sprint(pageListeners[reaction.MessageID].CurrentPage) + "/" + fmt.Sprint(pageListeners[reaction.MessageID].PageLimit),
					},
				})
			}
			// end of the left arrow :sunglasses:
			break
		case rightArrow:
			// update last used so the listener isn't deemed unused and deleted
			pageListeners[reaction.MessageID].LastUsed = MakeTimestamp()
			// remove reaction for better user expirence
			session.MessageReactionRemove(reaction.ChannelID, reaction.MessageID, rightArrow, reaction.UserID)
			// if the page we're on right now is already the maximum page length we have, on page 7 out of 7
			if pageListeners[reaction.MessageID].PageLimit == pageListeners[reaction.MessageID].CurrentPage {
				break
			}
			// update current page by 1
			pageListeners[reaction.MessageID].CurrentPage++
			// Update embed to be +1 page if possible
			// the max, we get this by getting the current page starting at index 1 and timsing it by 10 to set a limit
			var max = pageListeners[reaction.MessageID].CurrentPage * 10
			// we subtract one, this is because arrays and slices start at index 0, the max assums that it starts at index 1
			max--
			// the min is just the max minus 9, this collects all 10 between the scope
			var min = max - 9
			// if the leaderboard is for artists
			if pageListeners[reaction.MessageID].Type == "artist" {
				// if a specific artist is mentioned, this is a completely different output then normal so we gotta do it fully differently
				if pageListeners[reaction.MessageID].Specific {
					// build string to send
					toSend := ConvertToSendAbleStringSpecific(pageListeners[reaction.MessageID].SpecificArtist, pageListeners[reaction.MessageID].SpecificArtistStruct, pageListeners[reaction.MessageID].Author, min, max)
					// edit embed with the string that's built
					session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
						Description: toSend,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "Page " + fmt.Sprint(pageListeners[reaction.MessageID].CurrentPage) + "/" + fmt.Sprint(pageListeners[reaction.MessageID].PageLimit),
						},
					})
				} else {
					// now we're on to just normal artists leaderboard, we build the string like normal for artists
					toSend, _ := ConvertToSendableString(pageListeners[reaction.MessageID].Data, "artist", min, max)
					// edit the embed to that string we built
					session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
						Description: toSend,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "Page " + fmt.Sprint(pageListeners[reaction.MessageID].CurrentPage) + "/" + fmt.Sprint(pageListeners[reaction.MessageID].PageLimit),
						},
					})
				}
			} else {
				// onto songs, same process build a string
				toSend, _ := ConvertToSendableString(pageListeners[reaction.MessageID].Data, "song", min, max)
				// edit embed to string we built
				session.ChannelMessageEditEmbed(reaction.ChannelID, reaction.MessageID, &discordgo.MessageEmbed{
					Description: toSend,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Page " + fmt.Sprint(pageListeners[reaction.MessageID].CurrentPage) + "/" + fmt.Sprint(pageListeners[reaction.MessageID].PageLimit),
					},
				})
			}
			// done :sunglasses:
			break
		case destroyEmoji:
			// remove the specific page listener from the map, no longer listening for reactions
			delete(pageListeners, reaction.MessageID)
			// delete the embed the bot made, just cleans itself up.
			session.ChannelMessageDelete(reaction.ChannelID, reaction.MessageID)
			// done :sunglasses:
			break
		default:
			// done :sunglasses:
			break
		}
	}
}

// commands are processed here.
func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// if the message has a content and the first character is the current prefix
	if len(message.Content) > 0 && strings.HasPrefix(message.Content, Prefix) {
		// init command and args
		var command string
		var args []string
		// ".top song" -> {".top", "song"}
		args = strings.Split(message.Content, " ")
		//{".top", "song"} -> "top"
		command = strings.Replace(args[0], Prefix, "", 1)
		//{".top", "song"} -> {"song"}
		args = args[1:]
		// "tOP" -> "top"
		command = strings.ToLower(command)
		// Top command for leaderboards
		if command == "top" {
			// If no arguments are defined, send the help embed for the top command
			if len(args) < 1 {
				session.ChannelMessageSendEmbed(message.ChannelID, topHelpEmbed)
				return
			}
			// If they're trying to do the top command for artists
			if strings.ToLower(args[0]) == "artist" {
				// If they aren't in the database!
				if _, ok := mainMap[message.Author.ID]; !ok {
					session.ChannelMessageSend(message.ChannelID, "No data for you is available!")
					return
				}
				// If there's a specific artist defined we assume that they want the leaderboard for most listened songs for that artist
				if len(args) > 1 {
					//pull the specified artist
					var specificArtist string
					// {"song", "Lil", "Peep"} -> {"Lil", "Peep"}
					args = args[1:]
					// {"Lil", "Peep"} -> "Lil Peep"
					specificArtist = strings.Join(args, " ")
					// if the specified artist isn't in their artist database
					if _, ok := mainMap[message.Author.ID].ListeningArray[specificArtist]; !ok {
						session.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{Description: "**No data for " + specificArtist + " was found!** (case sensitive)"})
						return
					}
					// convert the map to a json struct then give us back the specific artist out of the artist slice
					jsonStructOfUserSpecificData := ConvertMapDataToStructArtistSpecific(specificArtist, mainMap[message.Author.ID])
					// Convert the json object to a string
					toSend := ConvertToSendAbleStringSpecific(specificArtist, jsonStructOfUserSpecificData, message.Author.Username+"#"+message.Author.Discriminator, 0, 9)
					// getting the page limit (this was a pain in my ass)
					var pageLimit = int(math.Ceil(float64(len(jsonStructOfUserSpecificData.Songs)) / 10.0))
					// build embed to send
					var embed = &discordgo.MessageEmbed{
						Description: toSend,
						Footer: &discordgo.MessageEmbedFooter{
							Text: "Page 1/" + fmt.Sprint(pageLimit),
						},
					}
					//send embed
					msg, err := session.ChannelMessageSendEmbed(message.ChannelID, embed)
					if err != nil {
						return
					}
					// add reactions for page listener
					session.MessageReactionAdd(msg.ChannelID, msg.ID, leftArrow)
					session.MessageReactionAdd(msg.ChannelID, msg.ID, rightArrow)
					session.MessageReactionAdd(msg.ChannelID, msg.ID, destroyEmoji)
					// data is the users map data converted to a struct (should be un sorted)
					data := ConvertMapDataToStruct(mainMap[message.Author.ID])
					// add to the page listeners
					pageListeners[msg.ID] = &ReactionListener{
						Type:                 "artist",
						Specific:             true,
						CurrentPage:          1,
						PageLimit:            pageLimit,
						SpecificArtist:       specificArtist,
						Data:                 data,
						Author:               message.Author.Username + "#" + message.Author.Discriminator,
						SpecificArtistStruct: jsonStructOfUserSpecificData,
						UserID:               message.Author.ID,
						LastUsed:             MakeTimestamp(),
					}
					// done, we return so it wont continue to the below code, that would be a shit show lol
					return
				}
				// Here we get the leaderboard for the most listened artists
				// this is the users map data
				var userSpecificData = mainMap[message.Author.ID]
				// converting the users map data to a struct, this struct has sorted data
				jsonStructOfUserSpecificData := ConvertMapDataToStruct(userSpecificData)
				// convert struct to string to send in embed
				toSend, num := ConvertToSendableString(jsonStructOfUserSpecificData, "artist", 0, 9)
				// p a i n. i n. m y. a s s.
				var pageLimit = int(math.Ceil(float64(num) / 10.0))
				// build embed to send
				var embed = &discordgo.MessageEmbed{
					Description: toSend,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Page 1/" + fmt.Sprint(pageLimit),
					},
				}
				// send the embed
				msg, err := session.ChannelMessageSendEmbed(message.ChannelID, embed)
				if err != nil {
					return
				}
				// add reactions to sent embed
				session.MessageReactionAdd(msg.ChannelID, msg.ID, leftArrow)
				session.MessageReactionAdd(msg.ChannelID, msg.ID, rightArrow)
				session.MessageReactionAdd(msg.ChannelID, msg.ID, destroyEmoji)
				// add to the page listeners
				pageListeners[msg.ID] = &ReactionListener{
					Type:        "artist",
					Specific:    false,
					CurrentPage: 1,
					PageLimit:   pageLimit,
					Data:        jsonStructOfUserSpecificData,
					Author:      message.Author.Username + "#" + message.Author.Discriminator,
					UserID:      message.Author.ID,
					LastUsed:    MakeTimestamp(),
				}
				// donezo
			} else if strings.ToLower(args[0]) == "song" {
				// here is where we do the song leader board
				// check if the user exists in the map database
				if _, ok := mainMap[message.Author.ID]; !ok {
					session.ChannelMessageSend(message.ChannelID, "No data for you is available!")
					return
				}
				// the users map data
				var userSpecificData = mainMap[message.Author.ID]
				// converts users map data to struct
				jsonStructOfUserSpecificData := ConvertMapDataToStruct(userSpecificData)
				// build string to send in embed
				toSend, num := ConvertToSendableString(jsonStructOfUserSpecificData, "song", 0, 9)
				// again. pain.
				var pageLimit = int(math.Ceil(float64(num) / 10.0))
				// build embed
				var embed = &discordgo.MessageEmbed{
					Description: toSend,
					Footer: &discordgo.MessageEmbedFooter{
						Text: "Page 1/" + fmt.Sprint(pageLimit),
					},
				}
				// send embed
				msg, err := session.ChannelMessageSendEmbed(message.ChannelID, embed)
				if err != nil {
					return
				}
				// add reactions
				session.MessageReactionAdd(msg.ChannelID, msg.ID, leftArrow)
				session.MessageReactionAdd(msg.ChannelID, msg.ID, rightArrow)
				session.MessageReactionAdd(msg.ChannelID, msg.ID, destroyEmoji)
				// add to the page listeners
				pageListeners[msg.ID] = &ReactionListener{
					Type:        "song",
					Specific:    false,
					CurrentPage: 1,
					PageLimit:   pageLimit,
					Data:        jsonStructOfUserSpecificData,
					Author:      message.Author.Username + "#" + message.Author.Discriminator,
					UserID:      message.Author.ID,
					LastUsed:    MakeTimestamp(),
				}
				// DONE :SUNGLASSES
			} else {
				// if the option is anything other than "artist" or "song" then send the help embed
				session.ChannelMessageSendEmbed(message.ChannelID, topHelpEmbed)
			}
		} else if command == "help" {
			// help command, displays very useful information
			var embed = &discordgo.MessageEmbed{
				Description: "This bot is for tracking discord spotify \"listening\" statuses to collect all songs you've listened to, sorted by artist and times played.\n\nPersonal command: **__top__**, run the command for more information!\n\nOther users command: **__utop__**, run the command for more information!",
				Footer: &discordgo.MessageEmbedFooter{
					Text: "github.com/postrequest69",
				},
			}
			// send help embed.
			session.ChannelMessageSendEmbed(message.ChannelID, embed)
		}
	}
}

// inits bot
func main() {
	bot, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session")
		return
	}
	bot.Identify.Intents = discordgo.IntentsGuildPresences | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMessageReactions
	bot.AddHandler(ready)
	bot.AddHandler(messageCreate)
	bot.AddHandler(presenceListen)
	bot.AddHandler(reactionListen)
	// this part keeps the bot online and makes sure when it exits the database will save (the 60 second saves will be useless in the case of a crash, this will be useful)
	err = bot.Open()
	if err != nil {
		fmt.Println("error opening connection,")
		return
	}
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
	_ = bot.Close()
	toSave := ConvertMapToJSONStruct(mainMap)
	toSaveString, _ := json.Marshal(toSave)
	UpdateDatabase(string(toSaveString))
}
