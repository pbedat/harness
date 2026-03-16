package cli

import (
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"tailscale.com/tsnet"

	"github.com/pbedat/harness/modules/email/adapters/in/postmark"
	"github.com/pbedat/harness/modules/email/app"
)

func newServeCmd(application *app.Application) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Start the email server",
		Example: `  email serve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

			e := echo.New()
			e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
				LogURI:      true,
				LogStatus:   true,
				LogMethod:   true,
				LogLatency:  true,
				HandleError: true,
				LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
					evt := logger.Info()
					if v.Error != nil {
						evt = logger.Error().Err(v.Error)
					}
					evt.
						Str("method", v.Method).
						Str("uri", v.URI).
						Int("status", v.Status).
						Dur("latency", v.Latency).
						Msg("request")
					return nil
				},
			}))
			postmark.Register(application, e, viper.GetString("postmark_user"), viper.GetString("postmark_pass"))

			srv := &tsnet.Server{
				Hostname: viper.GetString("ts_hostname"),
			}
			defer srv.Close()

			ln, err := srv.ListenFunnel("tcp", ":443")
			if err != nil {
				return err
			}
			defer ln.Close()

			domains := srv.CertDomains()
			if len(domains) > 0 {
				log.Printf("serving on https://%s", domains[0])
			} else {
				log.Println("serving on tailscale funnel")
			}

			go func() {
				if err := http.Serve(ln, e); err != nil && err != http.ErrServerClosed {
					log.Fatalf("server error: %v", err)
				}
			}()

			<-ctx.Done()
			log.Println("shutting down")
			return nil
		},
	}
}
