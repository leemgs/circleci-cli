package cmd

import (
	"os"
	"path"

	"github.com/CircleCI-Public/circleci-cli/logger"
	"github.com/CircleCI-Public/circleci-cli/settings"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var defaultEndpoint = "https://circleci.com/graphql-unstable"

// rootCmd is used internally and global to the package but not exported
// therefore we can use it in other commands, like `usage`
// it should be set once when Execute is first called
var rootCmd *cobra.Command

// Execute adds all child commands to rootCmd and
// sets flags appropriately. This function is called
// by main.main(). It only needs to happen once to
// the rootCmd.
func Execute() {
	command := MakeCommands()
	if err := command.Execute(); err != nil {
		os.Exit(-1)
	}
}

// Logger is exposed here so we can access it from subcommands.
// This allows us to print to the log at anytime from within the `cmd` package.
var Logger *logger.Logger

func hasAnnotations(cmd *cobra.Command) bool {
	return len(cmd.Annotations) > 0
}

var usageTemplate = `
Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}
Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if (HasAnnotations .)}}

Args:{{range $arg, $desc := .Annotations}}
  {{rpad $arg 11}} {{$desc}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}
Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}`

// MakeCommands creates the top level commands
func MakeCommands() *cobra.Command {
	rootCmd = &cobra.Command{
		Use:   "circleci",
		Short: "Use CircleCI from the command line.",
		Long:  `Use CircleCI from the command line.`,
	}

	// For supporting "Args" in command usage help
	cobra.AddTemplateFunc("HasAnnotations", hasAnnotations)
	rootCmd.SetUsageTemplate(usageTemplate)

	rootCmd.AddCommand(newTestsCommand())
	rootCmd.AddCommand(newQueryCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newOrbCommand())
	rootCmd.AddCommand(newLocalCommand())
	rootCmd.AddCommand(newBuildCommand())
	rootCmd.AddCommand(newVersionCommand())
	rootCmd.AddCommand(newDiagnosticCommand())
	rootCmd.AddCommand(newSetupCommand())
	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newNamespaceCommand())
	rootCmd.AddCommand(newUsageCommand())
	rootCmd.AddCommand(newStepCommand())
	rootCmd.PersistentFlags().Bool("verbose", false, "Enable verbose logging.")
	rootCmd.PersistentFlags().String("endpoint", defaultEndpoint, "the endpoint of your CircleCI GraphQL API")
	rootCmd.PersistentFlags().String("token", "", "your token for using CircleCI")

	for _, flag := range []string{"endpoint", "token", "verbose"} {
		bindCobraFlagToViper(rootCmd, flag)
	}

	// Cobra has a peculiar default behaviour:
	// https://github.com/spf13/cobra/issues/340
	// If you expose a command with `RunE`, and return an error from your
	// command, then Cobra will print the error message, followed by the usage
	// information for the command. This makes it really difficult to see what's
	// gone wrong. It usually prints a one line error message followed by 15
	// lines of usage information.
	// This flag disables that behaviour, so that if a comment fails, it prints
	// just the error message.
	rootCmd.SilenceUsage = true

	return rootCmd
}

func bindCobraFlagToViper(command *cobra.Command, flag string) {
	if err := viper.BindPFlag(flag, command.PersistentFlags().Lookup(flag)); err != nil {
		panic(errors.Wrapf(err, "internal error binding cobra flag '%s' to viper", flag))
	}
}

func init() {
	cobra.OnInitialize(prepare)

	configDir := path.Join(settings.UserHomeDir(), ".circleci")

	viper.SetConfigName("cli")
	viper.AddConfigPath(configDir)
	viper.SetEnvPrefix("circleci_cli")
	viper.AutomaticEnv()

	if err := settings.EnsureSettingsFileExists(configDir, "cli.yml"); err != nil {
		panic(err)
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func prepare() {
	Logger = logger.NewLogger(viper.GetBool("verbose"))
}
