package main

import (
	"encoding/json"
	"github.com/ChimeraCoder/anaconda"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-martini/martini"
	_ "github.com/joho/godotenv/autoload"
	"github.com/martini-contrib/render"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {

	m := martini.Classic()
	m.Use(render.Renderer())

	m.Use(func(res http.ResponseWriter, req *http.Request) {
		res.Header().Set("Access-Control-Allow-Origin", "*")
	})

	m.Get("/status", func() string {
		return "up"
	})

	m.Get("/:screen_name", func(params martini.Params, r render.Render) {
		sn := params["screen_name"]
		mc := memcache.New(os.Getenv("MEMCACHE_SERVER"))

		anaconda.SetConsumerKey(os.Getenv("TWITTER_KEY"))
		anaconda.SetConsumerSecret(os.Getenv("TWITTER_SECRET"))

		t, err := mc.Get(sn)

		if err == memcache.ErrCacheMiss {
			err = nil

			api := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

			v := url.Values{}
			v.Set("screen_name", sn)
			v.Set("count", "5")

			tweets, err := api.GetUserTimeline(v)

			if err != nil {
				newmap := map[string]interface{}{"response": err}
				r.JSON(400, newmap)
			}

			json, err := json.Marshal(tweets)

			if err != nil {
				newmap := map[string]interface{}{"response": err}
				r.JSON(400, newmap)
			}

			log.Printf("Saving %s's tweets to memcache", sn)

			err = mc.Set(&memcache.Item{Key: sn, Value: json, Expiration: 900})

			if err != nil {
				newmap := map[string]interface{}{"response": err}
				r.JSON(400, newmap)
			}

			newmap := map[string]interface{}{"response": tweets}
			r.JSON(200, newmap)

		} else {
			log.Printf("Found %s's tweets in memcache", sn)

			var tw []anaconda.Tweet
			err := json.Unmarshal(t.Value, &tw)

			if err == nil {
				newmap := map[string]interface{}{"response": tw}
				r.JSON(200, newmap)
			} else {
				newmap := map[string]interface{}{"response": err}
				r.JSON(400, newmap)
			}

		}

	})

	m.Run()
}
