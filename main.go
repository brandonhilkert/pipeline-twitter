package main

import (
	"encoding/json"
	"github.com/ChimeraCoder/anaconda"
	"github.com/go-martini/martini"
	_ "github.com/joho/godotenv/autoload"
	"github.com/martini-contrib/gorelic"
	"github.com/martini-contrib/render"
	"github.com/pmylund/go-cache"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var (
	c = cache.New(15*time.Minute, 1*time.Minute)
)

func main() {
	m := martini.Classic()
	m.Use(render.Renderer())

	gorelic.InitNewrelicAgent(os.Getenv("NEWRELIC_LICENSE_KEY"), "PipelineTwitter", true)
	m.Use(gorelic.Handler)

	m.Use(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", "*")
	})

	m.Get("/status", func() string {
		return "up"
	})

	m.Get("/favicon.ico", func() string {
		return "nope"
	})

	m.Get("/:screen_name", func(params martini.Params, r render.Render) {
		screenName := params["screen_name"]

		if t, found := c.Get(screenName); found {
			var tw []anaconda.Tweet

			log.Println("CACHE HIT: Found tweets for", screenName)
			err := json.Unmarshal(t.([]byte), &tw)

			if err == nil {
				newmap := map[string]interface{}{"response": tw}
				r.JSON(200, newmap)
			} else {
				newmap := map[string]interface{}{"response": err}
				r.JSON(400, newmap)
			}

			return
		}

		anaconda.SetConsumerKey(os.Getenv("TWITTER_KEY"))
		anaconda.SetConsumerSecret(os.Getenv("TWITTER_SECRET"))

		api := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

		v := url.Values{}
		v.Set("screen_name", screenName)
		v.Set("count", "5")

		log.Println("CACHE MISS: Getting tweets for", screenName)
		tweets, err := api.GetUserTimeline(v)

		go func() {
			j, err := json.Marshal(tweets)
			if err != nil {
				log.Println("Error marshaling tweets:", err)
			}

			c.Set(screenName, j, 0)
			log.Println("Saved tweets for:", screenName)
			log.Println("Total number of cached users:", c.ItemCount())
		}()

		if err == nil {
			newmap := map[string]interface{}{"response": tweets}
			r.JSON(200, newmap)
		} else {
			newmap := map[string]interface{}{"response": err}
			r.JSON(400, newmap)
		}
	})

	m.Run()
}
