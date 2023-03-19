package main

import (
	"context"
	ngrokPlugin "github.com/iamelevich/pocketbase-plugin-ngrok"
	"log"

	"github.com/pocketbase/pocketbase"
)

func main() {
	app := pocketbase.New()

	// Setup ngrok
	ngrokPlugin.MustRegister(app, &ngrokPlugin.Options{
		Ctx:       context.Background(),
		Enabled:   true,
		AuthToken: "YOUR_NGROK_AUTH_TOKEN", // Better to use ENV variable for that
		AfterSetup: func(url string) error {
			log.Printf("Started ngrok tunnel at %s", url)
			return nil
		},
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
