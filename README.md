
![Coverage](https://img.shields.io/badge/Coverage-21.1%25-red)
<!-- TOC -->
  * [Overview](#overview)
    * [Requirements](#requirements)
    * [Installation](#installation)
    * [Example](#example)
  * [Contributing](#contributing)
<!-- TOC -->

## Overview

This plugin allow expose local [Pocketbase](https://github.com/pocketbase/pocketbase) with [ngrok](https://ngrok.com/)

This plugin can be used for development purposes, when you need to expose your local Pocketbase instance to the internet. For example, you can use it to test your Pocketbase app on mobile device.

### Requirements

- Go 1.18+
- [Pocketbase](https://github.com/pocketbase/pocketbase) 0.13+

### Installation

```bash
go get github.com/iamelevich/pocketbase-plugin-ngrok
```

### Example

You can check examples in [examples folder](/examples)

```go
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
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
```

## Contributing

This pocketbase plugin is free and open source project licensed under the [MIT License](LICENSE.md).
You are free to do whatever you want with it, even offering it as a paid service.
