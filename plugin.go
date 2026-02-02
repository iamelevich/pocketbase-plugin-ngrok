package pocketbase_plugin_ngrok

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/fatih/color"
	"github.com/pocketbase/pocketbase/core"
	"golang.ngrok.com/ngrok/v2"
)

// Options defines optional struct to customize the default plugin behavior.
type Options struct {
	// Ctx is a context that will be used to start ngrok tunnel.
	Ctx context.Context

	// Enabled defines if ngrok tunnel should be started.
	Enabled bool

	// Enable logging of ngrok events to pocketbase logger
	EnableLogging bool

	// AuthToken is your ngrok auth token. You can get it from https://dashboard.ngrok.com/auth
	AuthToken string

	// AfterSetup is a callback function that will be called after ngrok tunnel is started.
	AfterSetup func(url *url.URL) error
}

type Plugin struct {
	// app is a Pocketbase application instance.
	app core.App

	// options is a plugin options.
	options *Options
}

// Validate plugin options. Return error if some option is invalid.
func (p *Plugin) Validate() error {
	if p.options == nil {
		return fmt.Errorf("options is required")
	}

	if p.options.Ctx == nil {
		return fmt.Errorf("context is required")
	}

	if p.app == nil {
		return fmt.Errorf("app is required")
	}

	if p.options.Enabled && p.options.AuthToken == "" {
		return fmt.Errorf("AuthToken is required when ngrok is enabled")
	}

	return nil
}

func (p *Plugin) exposeNgrok(e *core.ServeEvent) error {
	if p.options.Enabled {
		var agent ngrok.Agent
		var agentErr error
		if p.options.EnableLogging {
			agent, agentErr = ngrok.NewAgent(
				ngrok.WithAuthtoken(p.options.AuthToken),
				ngrok.WithLogger(e.App.Logger()),
			)
		} else {
			agent, agentErr = ngrok.NewAgent(
				ngrok.WithAuthtoken(p.options.AuthToken),
			)
		}
		if agentErr != nil {
			return agentErr
		}

		tun, err := agent.Forward(
			p.options.Ctx,
			ngrok.WithUpstream("tcp://"+e.Server.Addr),
		)
		if err != nil {
			return err
		}

		if p.options.AfterSetup != nil {
			if afterErr := p.options.AfterSetup(tun.URL()); afterErr != nil {
				return afterErr
			}
		}

		date := new(strings.Builder)
		log.New(date, "", log.LstdFlags).Print()

		bold := color.New(color.Bold).Add(color.FgGreen)
		_, _ = bold.Printf(
			"%s Ngrok tunnel started at %s\n",
			strings.TrimSpace(date.String()),
			color.CyanString("%s", tun.URL()),
		)

		regular := color.New()
		_, _ = regular.Printf(" ➜ REST API: %s\n", color.CyanString("%s/api/", tun.URL()))
		_, _ = regular.Printf(" ➜ Admin UI: %s\n", color.CyanString("%s/_/", tun.URL()))
	}
	return nil
}

// MustRegister is a helper function that registers plugin and panics if error occurred.
func MustRegister(app core.App, options *Options) *Plugin {
	if p, err := Register(app, options); err != nil {
		panic(err)
	} else {
		return p
	}
}

// Register registers plugin.
func Register(app core.App, options *Options) (*Plugin, error) {
	p := &Plugin{app: app}

	// Set default options
	if options != nil {
		p.options = options
	} else {
		p.options = &Options{}
	}

	// Validate options
	if err := p.Validate(); err != nil {
		return p, err
	}

	app.OnServe().BindFunc(func(se *core.ServeEvent) error {
		if err := p.exposeNgrok(se); err != nil {
			return err
		}
		return se.Next()
	})

	return p, nil
}
