package main

import (
	"context"
	"log"
	"net/url"

	ngrokPlugin "github.com/iamelevich/pocketbase-plugin-ngrok"

	"github.com/pocketbase/pocketbase"
)

func main() {
	app := pocketbase.New()

	// Setup ngrok
	ngrokPlugin.MustRegister(app, &ngrokPlugin.Options{
		Ctx:       context.Background(),
		Enabled:   true,
		AuthToken: "YOUR_NGROK_AUTH_TOKEN", // Better to use ENV variable for that
		AfterSetup: func(url *url.URL) error {
			log.Printf("Started ngrok tunnel at %s", url.String())
			return nil
		},
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
