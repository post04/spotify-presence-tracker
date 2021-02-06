package main

import (
	"fmt"
	"os"
	"sort"
	"time"
)

// ConvertMapToJSONStruct - Converst the map data to a JSON struct, this is used to save the data to the database.json :D
func ConvertMapToJSONStruct(data map[string]*MainData) Database {
	var toReturn Database
	for _, d := range data {
		var main MainStruct
		var mainListen []ListenArrayStruct
		for _, listen := range d.ListeningArray {
			var listenStruct ListenArrayStruct
			listenStruct.ArtistName = listen.ArtistName
			listenStruct.TimesListend = listen.TimesListend
			var mainSongs []SongsStruct
			//make songs array
			for _, songUwU := range listen.Songs {
				var songStruct SongsStruct
				songStruct.SongName = songUwU.SongName
				songStruct.TimesListened = songUwU.TimesListened
				songStruct.Artist = listen.ArtistName
				mainSongs = append(mainSongs, songStruct)
			}
			listenStruct.Songs = mainSongs
			// append artist to artist struct array
			mainListen = append(mainListen, listenStruct)
		}
		//append artists struct array
		main.Userid = d.Userid
		main.User = d.User
		main.ListeningArray = mainListen
		toReturn = append(toReturn, main)
	}
	return toReturn
}

// UpdateDatabase - Updates the main database.json file.
func UpdateDatabase(tosave string) {
	f, err := os.Create("database.json")
	if err != nil {
		fmt.Println("error saving Database!")
		f.Close()
		return
	}
	f.WriteString(tosave)
	f.Close()
	fmt.Println("Saved Database!")
}

// ConverToMap - Convert the json data to a map we can use.
func ConverToMap(data Database) map[string]*MainData {
	allData := make(map[string]*MainData)
	for _, normalData := range data {
		//if user doesn't exist
		if _, ok := allData[normalData.Userid]; !ok {
			listenArr := make(map[string]*Artist)
			for _, listen := range normalData.ListeningArray {
				//artist doesn't exist in lsitening array
				if _, ok := listenArr[listen.ArtistName]; !ok {
					songsArr := make(map[string]*Song)
					for _, songData := range listen.Songs {
						// Song doesn't exist in songs array
						if _, ok := songsArr[songData.SongName]; !ok {
							songsArr[songData.SongName] = &Song{
								SongName:      songData.SongName,
								TimesListened: songData.TimesListened,
							}
						}
					}
					listenArr[listen.ArtistName] = &Artist{
						ArtistName:   listen.ArtistName,
						Songs:        songsArr,
						TimesListend: listen.TimesListend,
					}
				}
			}
			allData[normalData.Userid] = &MainData{
				Userid:         normalData.Userid,
				User:           normalData.User,
				ListeningArray: listenArr,
			}
		}
	}
	return allData
}

// MakeTimestamp - makes a ms timestamp
func MakeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// ConvertMapDataToStruct - Convert the main map to a json struct, this is used to sort user data on fetch. (artist/general)
func ConvertMapDataToStruct(d *MainData) MainStruct {
	var main MainStruct
	var mainListen []ListenArrayStruct
	for _, listen := range d.ListeningArray {
		var listenStruct ListenArrayStruct
		listenStruct.ArtistName = listen.ArtistName
		listenStruct.TimesListend = listen.TimesListend
		var mainSongs []SongsStruct
		//make songs array
		for _, songUwU := range listen.Songs {
			var songStruct SongsStruct
			songStruct.SongName = songUwU.SongName
			songStruct.TimesListened = songUwU.TimesListened
			songStruct.Artist = listen.ArtistName
			mainSongs = append(mainSongs, songStruct)
		}
		listenStruct.Songs = mainSongs
		// append artist to artist struct array
		mainListen = append(mainListen, listenStruct)
	}
	//append artists struct array
	main.Userid = d.Userid
	main.User = d.User
	main.ListeningArray = mainListen
	return main
}

// ConvertMapDataToStructArtistSpecific - Convert the main map to a json struct, this is used to sort user data on fetch. (song/specific)
func ConvertMapDataToStructArtistSpecific(artist string, d *MainData) ListenArrayStruct {
	var toReturn ListenArrayStruct
	toReturn.ArtistName = d.ListeningArray[artist].ArtistName
	toReturn.TimesListend = d.ListeningArray[artist].TimesListend
	var mainSongs []SongsStruct
	for _, songUwU := range d.ListeningArray[artist].Songs {
		var songStruct SongsStruct
		songStruct.SongName = songUwU.SongName
		songStruct.TimesListened = songUwU.TimesListened
		songStruct.Artist = toReturn.ArtistName
		mainSongs = append(mainSongs, songStruct)
	}
	toReturn.Songs = mainSongs
	return toReturn
}

// ConvertToSendableString - Convert the struct to a string to send. (not specific)
/*
lets just walk through this here, comments take a lot of effort
1.) add 1 to the maximum, I'm not sure why but it just works. (0 -> first part but 9 -> 9th part :what:)
2.) we check the mode, if it's song we do this
	a.) initiate an array of every song our user has listned to
	b.) sort
	c.) grab only the songs we need for the page number
	d.) build the string and return it
3.) we check the mode, if it's artist do this
	a.) initiate an array of every artist our user has listened to
	b.) sort
	c.) grab only what we need for the page number
	d.) build string and return
*/
func ConvertToSendableString(structData MainStruct, mode string, min, max int) (string, int) {
	//min++
	max++
	var toReturn string

	if mode == "song" {
		var songs []SongsStruct
		for _, artistlol := range structData.ListeningArray {
			for _, songlol := range artistlol.Songs {
				songs = append(songs, songlol)
			}
		}
		sort.Slice(songs, func(i, j int) bool {
			return songs[i].TimesListened > songs[j].TimesListened
		})
		var numreturn = len(songs)
		if max > numreturn {
			max = numreturn
		}
		fmt.Println(min, max, len(songs))
		songs = songs[min:max]
		toReturn += fmt.Sprintf("**Top songs for %s**\n\n", structData.User)
		var i = min + 1
		for _, xd := range songs {
			toReturn += fmt.Sprintf("%s.) **%s** by **%s** listened to %s time(s)\n", fmt.Sprint(i), xd.SongName, xd.Artist, fmt.Sprint(xd.TimesListened))
			i++
		}
		return toReturn, numreturn

	} else if mode == "artist" {

		var artists []ListenArrayStruct
		artists = structData.ListeningArray
		sort.Slice(artists, func(i, j int) bool {
			return artists[i].TimesListend > artists[j].TimesListend
		})
		var numreturn = len(artists)
		if max > numreturn {
			max = numreturn
		}
		artists = artists[min:max]
		toReturn += fmt.Sprintf("**Top artists for %s**\n\n", structData.User)
		var i = min + 1
		for _, xd := range artists {
			toReturn += fmt.Sprintf("%s.) **%s** listened to %s time(s)\n", fmt.Sprint(i), xd.ArtistName, fmt.Sprint(xd.TimesListend))
			i++
		}
		return toReturn, numreturn
	}
	return "No valid data!", 0
}

// ConvertToSendAbleStringSpecific - Convert the struct to a string to send. (specific)
/*
1.) sort the songs
2.) grab what we need for the page number
3.) build string and return
*/
func ConvertToSendAbleStringSpecific(artist string, data ListenArrayStruct, user string, min, max int) string {
	max++
	var toReturn string
	if len(data.Songs) < 1 {
		toReturn = "**No data avalible for " + artist + "** (case sensitive)"
		return toReturn
	}
	var songs = data.Songs
	sort.Slice(songs, func(i, j int) bool {
		return songs[i].TimesListened > songs[j].TimesListened
	})
	if max > len(songs) {
		max = len(songs)
	}
	songs = songs[min:max]
	toReturn += fmt.Sprintf("**Top songs for %s by %s**\n\n", user, artist)

	var i = min + 1
	for _, xd := range songs {
		toReturn += fmt.Sprintf("%s.) **%s** by **%s** listened to %s time(s)\n", fmt.Sprint(i), xd.SongName, xd.Artist, fmt.Sprint(xd.TimesListened))
		i++
	}

	return toReturn
}
