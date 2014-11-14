package main

import (
	"github.com/ChimeraCoder/anaconda"
	"github.com/go-martini/martini"
	_ "github.com/joho/godotenv/autoload"
	"github.com/martini-contrib/render"
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

		anaconda.SetConsumerKey(os.Getenv("TWITTER_KEY"))
		anaconda.SetConsumerSecret(os.Getenv("TWITTER_SECRET"))

		api := anaconda.NewTwitterApi(os.Getenv("TWITTER_ACCESS_TOKEN"), os.Getenv("TWITTER_ACCESS_TOKEN_SECRET"))

		v := url.Values{}
		v.Set("screen_name", params["screen_name"])
		v.Set("count", "5")

		tweets, err := api.GetUserTimeline(v)

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
