package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

// Conditions stores the conditions of a particular place
type Conditions struct {
	Name    string
	Coord   Coordinates
	Weather []Weather
	Wind    Wind
	Main    Main
	Sys     Sys
}

type Coordinates struct {
	Lat float32
	Lon float32
}

type Weather struct {
	Main        string
	Description string
}

type Wind struct {
	Speed float32
}

type Main struct {
	Temp     float32
	Pressure float32
	Humidity float32
}

type Sys struct {
	Country string
}

func main() {
	token := flag.String("token", "", "API key of the telegram bot.")
	api := flag.String("api", "", "API key of the Open Weather Map API")

	flag.Parse()
	if *token == "" || *api == "" {
		printUsage()
		os.Exit(1)
	}

	bot, err := tb.NewBot(tb.Settings{
		Token:  *token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	bot.Handle("/start", func(m *tb.Message) {
		bot.Send(m.Sender, `Hello there !
		I am Weather Bot. I can display the weather conditions of a particular city.
		I was made with the help of Go.`)
	})

	bot.Handle("/help", func(m *tb.Message) {
		bot.Send(m.Sender, `Here is how I work !

							/weather <city name>
							
							Examples:
								/weather Mumbai
								/weather New York`)
	})

	bot.Handle("/weather", func(m *tb.Message) {
		place := strings.TrimSpace(m.Text[8:])
		if place == "" {
			bot.Send(m.Sender, "Enter the name of the place whose weather conditions you want. Ex: /weather Mumbai ")
			return
		}
		url := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s", place, *api)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal("NewRequest: ", err)
			return
		}

		client := &http.Client{}

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Do: ", err)
			return
		}
		if resp.StatusCode == 404 {
			bot.Send(m.Sender, `Oops! This is embarassing :( 
								I couldn't find the place.`)
			return
		}
		defer resp.Body.Close()
		var cond Conditions
		data, _ := ioutil.ReadAll(resp.Body)
		log.Println(string(data))
		if err := json.Unmarshal(data, &cond); err != nil {
			log.Println(err)
			return
		}

		output := fmt.Sprintf("%s, %s\n\n\tCo-ordinates \nLatitiude: %f\nLongitude: %f\n\n\tWeather\nMain: %s\nDescription: %s\nTemperature: %f\nHumidity: %f\nWind Speed: %f",
			cond.Name, cond.Sys.Country, cond.Coord.Lat, cond.Coord.Lon, cond.Weather[0].Main, cond.Weather[0].Description, cond.Main.Temp, cond.Main.Humidity, cond.Wind.Speed)
		bot.Send(m.Sender, output)
	})

	bot.Handle(tb.OnText, func(m *tb.Message) {
		bot.Send(m.Sender, "I don't understand the command. Sorry  :(")
	})

	bot.Start()
}

func printUsage() {
	fmt.Printf("Usage: %s [options]\n", os.Args[0])
	fmt.Println("Options:")
	fmt.Println("\t -token\t access token of the telegram bot")
	fmt.Println("\t -api\t API key of Open Weather Map")
}
