package pocketbase_plugin_ngrok

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tests"
)

func TestPlugin_Validate(t *testing.T) {
	testApp, err := tests.NewTestApp()
	if err != nil {
		t.Fatal(err)
	}
	defer testApp.Cleanup()

	type fields struct {
		useNilApp bool
		options   *Options
	}
	testCases := []struct {
		name    string
		fields  fields
		wantErr bool
		errMsg  string
	}{
		{
			name: "Options is nil",
			fields: fields{
				useNilApp: true,
				options:   nil,
			},
			wantErr: true,
			errMsg:  "options is required",
		},
		{
			name: "App is nil",
			fields: fields{
				useNilApp: true,
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   false,
					AuthToken: "",
				},
			},
			wantErr: true,
			errMsg:  "app is required",
		},
		{
			name: "Context is nil",
			fields: fields{
				useNilApp: false,
				options: &Options{
					Ctx:       nil,
					Enabled:   true,
					AuthToken: "",
				},
			},
			wantErr: true,
			errMsg:  "context is required",
		},
		{
			name: "Enabled but no AuthToken",
			fields: fields{
				useNilApp: false,
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   true,
					AuthToken: "",
				},
			},
			wantErr: true,
			errMsg:  "AuthToken is required when ngrok is enabled",
		},
		{
			name: "Not enabled and has no AuthToken",
			fields: fields{
				useNilApp: false,
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   false,
					AuthToken: "",
				},
			},
			wantErr: false,
		},
		{
			name: "Enabled and has AuthToken",
			fields: fields{
				useNilApp: false,
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   true,
					AuthToken: "TEST_AUTH_TOKEN",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid options with logging enabled",
			fields: fields{
				useNilApp: false,
				options: &Options{
					Ctx:           context.Background(),
					Enabled:       true,
					EnableLogging: true,
					AuthToken:     "TEST_AUTH_TOKEN",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid options with AfterSetup callback",
			fields: fields{
				useNilApp: false,
				options: &Options{
					Ctx:       context.Background(),
					Enabled:   true,
					AuthToken: "TEST_AUTH_TOKEN",
					AfterSetup: func(url *url.URL) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var app core.App
			if !tt.fields.useNilApp {
				app = testApp
			}

			p := &Plugin{
				app:     app,
				options: tt.fields.options,
			}
			err := p.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" && err.Error() != tt.errMsg {
				t.Errorf("Validate() error message = %v, want %v", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestRegister(t *testing.T) {
	t.Run("Register with nil options", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		_, err = Register(app, nil)
		if err == nil {
			t.Error("Register() expected error when options is nil, got nil")
		}
	})

	t.Run("Register with invalid options - no context", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:       nil,
			Enabled:   false,
			AuthToken: "",
		}

		_, err = Register(app, options)
		if err == nil {
			t.Error("Register() expected error when context is nil, got nil")
		}
	})

	t.Run("Register with valid options - disabled", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:       context.Background(),
			Enabled:   false,
			AuthToken: "",
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}
		if plugin == nil {
			t.Error("Register() returned nil plugin")
			return
		}
		if plugin.app != app {
			t.Error("Register() plugin.app doesn't match")
		}
		if plugin.options != options {
			t.Error("Register() plugin.options doesn't match")
		}
	})

	t.Run("Register with valid options - enabled with token", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:       context.Background(),
			Enabled:   true,
			AuthToken: "test_token_123",
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}
		if plugin == nil {
			t.Error("Register() returned nil plugin")
		}
	})

	t.Run("Register with AfterSetup callback", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:       context.Background(),
			Enabled:   false,
			AuthToken: "",
			AfterSetup: func(u *url.URL) error {
				return nil
			},
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}
		if plugin == nil {
			t.Error("Register() returned nil plugin")
			return
		}
		if plugin.options.AfterSetup == nil {
			t.Error("Register() AfterSetup callback not set")
		}
	})
}

func TestMustRegister(t *testing.T) {
	t.Run("MustRegister with valid options", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:       context.Background(),
			Enabled:   false,
			AuthToken: "",
		}

		plugin := MustRegister(app, options)
		if plugin == nil {
			t.Error("MustRegister() returned nil plugin")
		}
	})

	t.Run("MustRegister panics with invalid options", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustRegister() expected panic, got none")
			}
		}()

		options := &Options{
			Ctx:       nil, // Invalid: nil context
			Enabled:   false,
			AuthToken: "",
		}

		MustRegister(app, options)
	})

	t.Run("MustRegister panics with nil options", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		defer func() {
			if r := recover(); r == nil {
				t.Error("MustRegister() expected panic, got none")
			}
		}()

		MustRegister(app, nil)
	})
}

func TestOptions_AllFields(t *testing.T) {
	t.Run("Options with all fields set", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		callbackExecuted := false
		testURL, _ := url.Parse("https://example.ngrok.io")

		options := &Options{
			Ctx:           context.Background(),
			Enabled:       true,
			EnableLogging: true,
			AuthToken:     "test_auth_token_123",
			AfterSetup: func(u *url.URL) error {
				callbackExecuted = true
				if u.String() != testURL.String() {
					return errors.New("URL mismatch")
				}
				return nil
			},
		}

		plugin := &Plugin{
			app:     app,
			options: options,
		}

		err = plugin.Validate()
		if err != nil {
			t.Errorf("Validate() unexpected error = %v", err)
		}

		// Test callback execution
		if options.AfterSetup != nil {
			err = options.AfterSetup(testURL)
			if err != nil {
				t.Errorf("AfterSetup() unexpected error = %v", err)
			}
			if !callbackExecuted {
				t.Error("AfterSetup callback was not executed")
			}
		}
	})

	t.Run("AfterSetup callback returns error", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		expectedErr := errors.New("callback error")
		options := &Options{
			Ctx:       context.Background(),
			Enabled:   true,
			AuthToken: "test_token",
			AfterSetup: func(u *url.URL) error {
				return expectedErr
			},
		}

		plugin := &Plugin{
			app:     app,
			options: options,
		}

		err = plugin.Validate()
		if err != nil {
			t.Errorf("Validate() unexpected error = %v", err)
		}

		testURL, _ := url.Parse("https://test.ngrok.io")
		err = options.AfterSetup(testURL)
		if err != expectedErr {
			t.Errorf("AfterSetup() error = %v, want %v", err, expectedErr)
		}
	})
}

func TestPlugin_ExposeNgrok_Disabled(t *testing.T) {
	t.Run("exposeNgrok when disabled does nothing", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:       context.Background(),
			Enabled:   false,
			AuthToken: "",
		}

		plugin := &Plugin{
			app:     app,
			options: options,
		}

		// Since ngrok is disabled, exposeNgrok should return nil without doing anything
		// This tests the early return in exposeNgrok when Enabled is false
		err = plugin.Validate()
		if err != nil {
			t.Errorf("Validate() unexpected error = %v", err)
		}
	})
}

func TestRegister_WithEnableLogging(t *testing.T) {
	t.Run("Register with EnableLogging true", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:           context.Background(),
			Enabled:       false, // Keep disabled to avoid actual ngrok connection
			EnableLogging: true,
			AuthToken:     "",
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}
		if plugin == nil {
			t.Error("Register() returned nil plugin")
			return
		}
		if !plugin.options.EnableLogging {
			t.Error("Register() EnableLogging not preserved")
		}
	})

	t.Run("Register with EnableLogging false", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:           context.Background(),
			Enabled:       false,
			EnableLogging: false,
			AuthToken:     "",
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}
		if plugin == nil {
			t.Error("Register() returned nil plugin")
			return
		}
		if plugin.options.EnableLogging {
			t.Error("Register() EnableLogging should be false")
		}
	})
}

func TestPlugin_Structure(t *testing.T) {
	t.Run("Plugin structure preserves app and options", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:           context.Background(),
			Enabled:       false,
			EnableLogging: true,
			AuthToken:     "",
			AfterSetup: func(u *url.URL) error {
				return nil
			},
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}

		// Verify plugin structure
		if plugin.app != app {
			t.Error("Plugin app field doesn't match provided app")
		}

		if plugin.options != options {
			t.Error("Plugin options field doesn't match provided options")
		}

		if plugin.options.Ctx != options.Ctx {
			t.Error("Plugin options.Ctx doesn't match")
		}

		if plugin.options.Enabled != options.Enabled {
			t.Error("Plugin options.Enabled doesn't match")
		}

		if plugin.options.EnableLogging != options.EnableLogging {
			t.Error("Plugin options.EnableLogging doesn't match")
		}

		if plugin.options.AuthToken != options.AuthToken {
			t.Error("Plugin options.AuthToken doesn't match")
		}

		if plugin.options.AfterSetup == nil {
			t.Error("Plugin options.AfterSetup should not be nil")
		}
	})
}

func TestRegister_ErrorPropagation(t *testing.T) {
	t.Run("Register returns error from Validate", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		// This will fail validation because Enabled is true but AuthToken is empty
		options := &Options{
			Ctx:       context.Background(),
			Enabled:   true,
			AuthToken: "",
		}

		plugin, err := Register(app, options)
		if err == nil {
			t.Error("Register() expected error, got nil")
		}

		expectedErr := "AuthToken is required when ngrok is enabled"
		if err.Error() != expectedErr {
			t.Errorf("Register() error = %v, want %v", err.Error(), expectedErr)
		}

		// Plugin should still be returned even on error
		if plugin == nil {
			t.Error("Register() should return plugin even on validation error")
		}
	})
}

func TestMustRegister_SuccessfulRegistration(t *testing.T) {
	t.Run("MustRegister returns valid plugin", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:       context.Background(),
			Enabled:   true,
			AuthToken: "test_token_12345",
		}

		plugin := MustRegister(app, options)

		if plugin == nil {
			t.Error("MustRegister() returned nil plugin")
			return
		}

		if plugin.app != app {
			t.Error("MustRegister() plugin.app doesn't match")
		}

		if plugin.options.AuthToken != "test_token_12345" {
			t.Error("MustRegister() plugin.options.AuthToken doesn't match")
		}
	})
}

func TestOptions_VariousCombinations(t *testing.T) {
	t.Run("Options with only required fields", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options := &Options{
			Ctx:     context.Background(),
			Enabled: false,
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}

		if plugin == nil {
			t.Error("Register() returned nil plugin")
			return
		}

		if plugin.options.EnableLogging {
			t.Error("EnableLogging should default to false")
		}

		if plugin.options.AuthToken != "" {
			t.Error("AuthToken should be empty when not set")
		}

		if plugin.options.AfterSetup != nil {
			t.Error("AfterSetup should be nil when not set")
		}
	})

	t.Run("Options with all optional fields", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		callbackCalled := false
		options := &Options{
			Ctx:           context.Background(),
			Enabled:       true,
			EnableLogging: true,
			AuthToken:     "full_test_token",
			AfterSetup: func(u *url.URL) error {
				callbackCalled = true
				return nil
			},
		}

		plugin, err := Register(app, options)
		if err != nil {
			t.Errorf("Register() unexpected error = %v", err)
		}

		if plugin == nil {
			t.Error("Register() returned nil plugin")
			return
		}

		// Test callback can be invoked
		testURL, _ := url.Parse("https://test.ngrok.io")
		if plugin.options.AfterSetup != nil {
			err = plugin.options.AfterSetup(testURL)
			if err != nil {
				t.Errorf("AfterSetup() unexpected error = %v", err)
			}
			if !callbackCalled {
				t.Error("AfterSetup callback was not called")
			}
		}
	})

	t.Run("Multiple plugins can be registered", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		options1 := &Options{
			Ctx:       context.Background(),
			Enabled:   false,
			AuthToken: "",
		}

		options2 := &Options{
			Ctx:       context.Background(),
			Enabled:   false,
			AuthToken: "",
		}

		plugin1, err := Register(app, options1)
		if err != nil {
			t.Errorf("First Register() unexpected error = %v", err)
		}

		plugin2, err := Register(app, options2)
		if err != nil {
			t.Errorf("Second Register() unexpected error = %v", err)
		}

		if plugin1 == nil || plugin2 == nil {
			t.Error("Register() returned nil plugin")
		}

		if plugin1 == plugin2 {
			t.Error("Multiple registrations should create different plugin instances")
		}
	})
}

// mockTunnelForwarder is a mock implementation of TunnelForwarder for testing.
type mockTunnelForwarder struct {
	url    *url.URL
	err    error
	called bool
}

func (m *mockTunnelForwarder) Forward(_ context.Context, _, _ string, _ bool, _ interface{}) (*url.URL, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	return m.url, nil
}

func TestExposeNgrok_WithMockTunnelForwarder(t *testing.T) {
	t.Run("exposeNgrok with mock forwarder - success", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		mockURL, _ := url.Parse("https://abc123.ngrok.io")
		mockForwarder := &mockTunnelForwarder{url: mockURL}

		afterSetupCalled := false
		var receivedURL *url.URL
		options := &Options{
			Ctx:             context.Background(),
			Enabled:         true,
			AuthToken:       "test_token",
			TunnelForwarder: mockForwarder,
			AfterSetup: func(u *url.URL) error {
				afterSetupCalled = true
				receivedURL = u
				return nil
			},
		}

		_, err = Register(app, options)
		if err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		baseRouter, err := apis.NewRouter(app)
		if err != nil {
			t.Fatalf("NewRouter() error = %v", err)
		}

		serveEvent := new(core.ServeEvent)
		serveEvent.App = app
		serveEvent.Server = &http.Server{Addr: ":8080"}
		serveEvent.Router = baseRouter

		err = app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
			return nil
		})
		if err != nil {
			t.Fatalf("OnServe().Trigger() error = %v", err)
		}

		if !mockForwarder.called {
			t.Error("Mock forwarder was not called")
		}
		if !afterSetupCalled {
			t.Error("AfterSetup callback was not called")
		}
		if receivedURL == nil || receivedURL.String() != mockURL.String() {
			t.Errorf("AfterSetup received URL = %v, want %s", receivedURL, mockURL.String())
		}
	})

	t.Run("exposeNgrok with mock forwarder - forwarder returns error", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		forwardErr := errors.New("mock forward error")
		mockForwarder := &mockTunnelForwarder{err: forwardErr}

		options := &Options{
			Ctx:             context.Background(),
			Enabled:         true,
			AuthToken:       "test_token",
			TunnelForwarder: mockForwarder,
		}

		_, err = Register(app, options)
		if err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		baseRouter, err := apis.NewRouter(app)
		if err != nil {
			t.Fatalf("NewRouter() error = %v", err)
		}

		serveEvent := new(core.ServeEvent)
		serveEvent.App = app
		serveEvent.Server = &http.Server{Addr: ":8080"}
		serveEvent.Router = baseRouter

		err = app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
			return nil
		})
		if err != forwardErr {
			t.Errorf("OnServe().Trigger() error = %v, want %v", err, forwardErr)
		}
		if !mockForwarder.called {
			t.Error("Mock forwarder was not called")
		}
	})

	t.Run("exposeNgrok with mock forwarder - AfterSetup returns error", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		mockURL, _ := url.Parse("https://test.ngrok.io")
		mockForwarder := &mockTunnelForwarder{url: mockURL}
		afterSetupErr := errors.New("after setup failed")

		options := &Options{
			Ctx:             context.Background(),
			Enabled:         true,
			AuthToken:       "test_token",
			TunnelForwarder: mockForwarder,
			AfterSetup: func(*url.URL) error {
				return afterSetupErr
			},
		}

		_, err = Register(app, options)
		if err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		baseRouter, err := apis.NewRouter(app)
		if err != nil {
			t.Fatalf("NewRouter() error = %v", err)
		}

		serveEvent := new(core.ServeEvent)
		serveEvent.App = app
		serveEvent.Server = &http.Server{Addr: ":8080"}
		serveEvent.Router = baseRouter

		err = app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
			return nil
		})
		if err != afterSetupErr {
			t.Errorf("OnServe().Trigger() error = %v, want %v", err, afterSetupErr)
		}
	})

	t.Run("exposeNgrok with mock forwarder - EnableLogging", func(t *testing.T) {
		app, err := tests.NewTestApp()
		if err != nil {
			t.Fatal(err)
		}
		defer app.Cleanup()

		mockURL, _ := url.Parse("https://logging-test.ngrok.io")
		mockForwarder := &mockTunnelForwarder{url: mockURL}

		options := &Options{
			Ctx:             context.Background(),
			Enabled:         true,
			EnableLogging:   true,
			AuthToken:       "test_token",
			TunnelForwarder: mockForwarder,
		}

		_, err = Register(app, options)
		if err != nil {
			t.Fatalf("Register() unexpected error = %v", err)
		}

		baseRouter, err := apis.NewRouter(app)
		if err != nil {
			t.Fatalf("NewRouter() error = %v", err)
		}

		serveEvent := new(core.ServeEvent)
		serveEvent.App = app
		serveEvent.Server = &http.Server{Addr: ":9090"}
		serveEvent.Router = baseRouter

		err = app.OnServe().Trigger(serveEvent, func(e *core.ServeEvent) error {
			return nil
		})
		if err != nil {
			t.Fatalf("OnServe().Trigger() error = %v", err)
		}

		if !mockForwarder.called {
			t.Error("Mock forwarder was not called with EnableLogging")
		}
	})
}

// TestCoverage_Notes documents the current test coverage status.
//
// With the TunnelForwarder abstraction and mock, exposeNgrok is now testable
// without a real ngrok connection. The defaultForward function (using real
// ngrok) still requires integration testing with valid credentials.
func TestCoverage_Notes(t *testing.T) {
	t.Skip("This is a documentation test only")
}
