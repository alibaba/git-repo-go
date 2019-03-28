# Command line framework

As a command line tool, `git-repo` uses [cobra](https://github.com/spf13/cobra)
as its command line framework and use [viper](https://github.com/spf13/viper)
for config file and environment variables parsing.


# Root command

Root command is defined in file `cmd/root.go`.  The following persistent flags
are defined in root command and are available for all sub commands:

* --config <config-file> : location of config file, which defines default settings.
* --verbose, -v : verbose mode, and multple `-v` can be used to output more.
* --quiet, -q : be quiet, not show notice messages.
* --single : run `git-repo` in single repository mode. Only a few commands support it.


# Default settings

Default settings are defined in `config/config.go`, such as DefaultLogLevel
("warn" as default log level), DefailtLogRotate (20MB as the default rotate size).

    const (
    	DefaultConfigPath = ".git-repo"
    	DefaultLogRotate  = 20 * 1024 * 1024
    	DefaultLogLevel   = "warn"
    	...


# Config file

Default config file is `git-repo.yml` in HOME directory.  Settings in config
file can be used as default if not command line argument is provided.

Example of `git-repo.yml`:

    logfile: /var/log/git-repo.log
    loglevel: warn
    logrotate: 10240000


# Environment

Environment variables with prefix `GIT_REPO` can be used to provide default
value for git-repo settings, such as:

    GIT_REPO_VERBOSE=2
    GIT_REPO_LOGLEVEL=info


# Adding a sub command

To create a sub command, just wirte a go program in `cmd/<sub-cmd>.go`.
Because all sub commands share the same package name, functions for each
sub commands should use the name of the command as prefix of function names
to avoid conflicts.

* Define a cobra.Command struct, e.g.:

        // initCmd represents the init command
        var initCmd = &cobra.Command{
        	Use:   "init",
        	Short: "Initialize manifest repo in the current directory",
        	RunE: func(cmd *cobra.Command, args []string) error {
        		return initCmdRunE()
        	},
        }

* Implement the `<name>CmdRunE()` function, such as:

        func initCmdRunE() error {
        	...
        }

* Define its options in `init()` function, such as:

        func init() {
        	initCmd.Flags().StringVarP(&initOptions.ManifestURL,
        		"manifest-url",
        		"u",
        		"",
        		"manifest repository location")
        	initCmd.Flags().StringVarP(&initOptions.ManifestBranch,
        		"manifest-branch",
        		"b",
        		"",
        		"manifest branch or revision")

                ... ...

        	rootCmd.AddCommand(initCmd)
        }

See `cmd/init.go` as an example.


# Console output

[multi-log](https://github.com/jiangxin/multi-log) is used to handle console
output and log to file.

By default, the following messages will be printed on console:

    log.Fatal("fatal message...")
    log.Error("error message...")
    log.Warn("warn message...")
    log.Notef("note message...")
    log.Printf("hello, %s.", "world") // The same as Notef

If opiton `--verbose` is provided in command line, will show info level
message:

    log.Info("info message...")

If provide extra verbose mode (-v -v), will show both info and debug level
messages:

    log.Info("info message...")
    log.Debug("debug message...")

If triple verbose is provided, will also so trace level messages:

    log.Info("info message...")
    log.Debug("debug message...")
    log.Trace("trace message...")

For quilt mode (option `--quiet` is provided), note level messages (output of
log.Note and log.Print families) will be suppressed.

See test case: t0001.


# Log to file

If location of logfile is defined (by config file or by options), log messages
will also saved in logfile.

Loglevel is defined using `--loglevel` option or related setting in config file.
Loglevel is not affected by verbose mode (option `--verbose`) and quiet mode
(option `--quiet`).


# Error handling

Do not call `os.Exit()` or `log.Fatal()` in sub commands, just return an error.
Root command wraps error in a Response object, and return to main function.

It's annoying for cobra to print unnecessary command usages while showing
errors.  To suppress unnecessary command usage output, root command is
initialized with the field `SilenceUsage` set to true.

The main function only print command usages for userError. See: `cmd/helper.go`.
