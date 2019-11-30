package main

import (
	"fmt"
	"github.com/luoyayu/netease_go/api"
	"github.com/spf13/afero"
	"log"
	"os"
)

func exampleDownloadSong() {
	song := api.SongStruct{Id: 560079}
	if err := song.Download("", -1); err != nil {
		panic(err)
	}
}

func exampleDownloadSongs() {
	songs := api.SongsStruct{
		Songs: []api.SongStruct{
			{Id: 560079},
			{Id: 1387099649},
			{Id: 514774419},
		},
	}

	if err := songs.Download("", nil); err != nil {
		panic(err)
	}
}

func exampleGetSongDetail() {
	song := api.GetSongDetail(560079)
	fmt.Printf("%+v\n", song)
}

func exampleLoginCellphone() {
	_ = afero.NewOsFs().Remove("cookies")
	if err := api.LoginCellphone(os.Getenv("PHONE"), os.Getenv("PASSWD")); err != nil {
		log.Println(err)
	}
}

func main() {
	//exampleDownloadSong()
	//exampleDownloadSongs()
	//exampleGetSongDetail()
	exampleLoginCellphone()
}
