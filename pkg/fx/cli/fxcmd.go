package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/alanshaw/buff/pkg/config"
	"github.com/alanshaw/buff/pkg/fx/app"
	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap/zapcore"
)

var log = logging.Logger("pkg/cli")

func FXCommand(doFunc any) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Apply network presets before loading config, but only for flags that
		// weren't explicitly set
		if err := config.LoadPresets(); err != nil {
			return fmt.Errorf("loading presets: %w", err)
		}

		userCfg, err := config.Load[config.AppConfig]()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		appCfg, err := userCfg.ToAppConfig()
		if err != nil {
			return fmt.Errorf("parsing config: %w", err)
		}

		app := fx.New(
			fx.RecoverFromPanics(),

			// provide fx with our logger for its events logged at debug level.
			// any fx errors will still be logged at the error level.
			fx.WithLogger(func() fxevent.Logger {
				el := &fxevent.ZapLogger{Logger: log.Desugar()}
				el.UseLogLevel(zapcore.DebugLevel)
				return el
			}),

			fx.StopTimeout(time.Minute),

			app.CommonModules(appCfg),

			// provide the command and args as a dependency
			fx.Supply(cmd),
			fx.Supply(args),

			// invoke the command
			fx.Invoke(doFunc),

			// shutdown the app once the command is complete
			fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
				lc.Append(fx.Hook{
					OnStart: func(ctx context.Context) error {
						shutdowner.Shutdown()
						return nil
					},
				})
			}),
		)

		if err := app.Err(); err != nil {
			return fmt.Errorf("building app: %w", err)
		}

		app.Run()
		return nil
	}
}
