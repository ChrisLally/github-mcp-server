package main

import (
	"context"
	"fmt"
	"io"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/github/github-mcp-server/pkg/github"
	iolog "github.com/github/github-mcp-server/pkg/log"
	"github.com/github/github-mcp-server/pkg/translations"
	gogithub "github.com/google/go-github/v69/github"
	"github.com/mark3labs/mcp-go/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	githubv4 "github.com/shurcooL/githubv4"
)

var version = "version"
var commit = "commit"
var date = "date"

var (
	rootCmd = &cobra.Command{
		Use:     "server",
		Short:   "GitHub MCP Server",
		Long:    `A GitHub MCP server that handles various tools and resources.`,
		Version: fmt.Sprintf("%s (%s) %s", version, commit, date),
	}

	stdioCmd = &cobra.Command{
		Use:   "stdio",
		Short: "Start stdio server",
		Long:  `Start a server that communicates via standard input/output streams using JSON-RPC messages.`,
		Run: func(_ *cobra.Command, _ []string) {
			logFile := viper.GetString("log-file")
			readOnly := viper.GetBool("read-only")
			exportTranslations := viper.GetBool("export-translations")
			logger, err := initLogger(logFile)
			if err != nil {
				stdlog.Fatal("Failed to initialize logger:", err)
			}
			logCommands := viper.GetBool("enable-command-logging")
			cfg := runConfig{
				readOnly:           readOnly,
				logger:             logger,
				logCommands:        logCommands,
				exportTranslations: exportTranslations,
			}
			if err := runStdioServer(cfg); err != nil {
				stdlog.Fatal("failed to run stdio server:", err)
			}
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	// Add global flags that will be shared by all commands
	rootCmd.PersistentFlags().Bool("read-only", false, "Restrict the server to read-only operations")
	rootCmd.PersistentFlags().String("log-file", "", "Path to log file")
	rootCmd.PersistentFlags().Bool("enable-command-logging", false, "When enabled, the server will log all command requests and responses to the log file")
	rootCmd.PersistentFlags().Bool("export-translations", false, "Save translations to a JSON file")
	rootCmd.PersistentFlags().String("gh-host", "", "Specify the GitHub hostname (for GitHub Enterprise etc.)")

	// Bind flag to viper
	_ = viper.BindPFlag("read-only", rootCmd.PersistentFlags().Lookup("read-only"))
	_ = viper.BindPFlag("log-file", rootCmd.PersistentFlags().Lookup("log-file"))
	_ = viper.BindPFlag("enable-command-logging", rootCmd.PersistentFlags().Lookup("enable-command-logging"))
	_ = viper.BindPFlag("export-translations", rootCmd.PersistentFlags().Lookup("export-translations"))
	_ = viper.BindPFlag("gh-host", rootCmd.PersistentFlags().Lookup("gh-host"))

	// Add subcommands
	rootCmd.AddCommand(stdioCmd)
}

func initConfig() {
	// Initialize Viper configuration
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
}

func initLogger(outPath string) (*log.Logger, error) {
	if outPath == "" {
		return log.New(), nil
	}

	file, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := log.New()
	logger.SetLevel(log.DebugLevel)
	logger.SetOutput(file)

	return logger, nil
}

type runConfig struct {
	readOnly           bool
	logger             *log.Logger
	logCommands        bool
	exportTranslations bool
}

func runStdioServer(cfg runConfig) error {
	// Create app context
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Create GH client
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		cfg.logger.Fatal("GITHUB_PERSONAL_ACCESS_TOKEN not set")
	}
	
	// Create HTTP client with auth
	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Base: http.DefaultTransport,
			Source: oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: token},
			),
		},
	}

	// Create GitHub REST client
	ghClient := gogithub.NewClient(httpClient)
	ghClient.UserAgent = fmt.Sprintf("github-mcp-server/%s", version)

	// Create GitHub GraphQL client
	graphqlClient := githubv4.NewClient(httpClient)

	// Check GH_HOST env var first, then fall back to viper config
	host := os.Getenv("GH_HOST")
	if host == "" {
		host = viper.GetString("gh-host")
	}

	if host != "" {
		var err error
		ghClient, err = ghClient.WithEnterpriseURLs(host, host)
		if err != nil {
			return fmt.Errorf("failed to create GitHub client with host: %w", err)
		}
		graphqlClient = githubv4.NewEnterpriseClient(host+"/api/graphql", httpClient)
	}

	t, dumpTranslations := translations.TranslationHelper()

	getClient := func(_ context.Context) (*gogithub.Client, *githubv4.Client, error) {
		return ghClient, graphqlClient, nil // closing over clients
	}
	// Create
	ghServer := github.NewServer(getClient, version, cfg.readOnly, t)
	stdioServer := server.NewStdioServer(ghServer)

	stdLogger := stdlog.New(cfg.logger.Writer(), "stdioserver", 0)
	stdioServer.SetErrorLogger(stdLogger)

	if cfg.exportTranslations {
		// Once server is initialized, all translations are loaded
		dumpTranslations()
	}

	// Start listening for messages
	errC := make(chan error, 1)
	go func() {
		in, out := io.Reader(os.Stdin), io.Writer(os.Stdout)

		if cfg.logCommands {
			loggedIO := iolog.NewIOLogger(in, out, cfg.logger)
			in, out = loggedIO, loggedIO
		}

		errC <- stdioServer.Listen(ctx, in, out)
	}()

	// Output github-mcp-server string
	_, _ = fmt.Fprintf(os.Stderr, "GitHub MCP Server running on stdio\n")

	// Wait for shutdown signal
	select {
	case <-ctx.Done():
		cfg.logger.Infof("shutting down server...")
	case err := <-errC:
		if err != nil {
			return fmt.Errorf("error running server: %w", err)
		}
	}

	return nil
}

// Print verbose startup logs and diagnostic information
func main() {
	fmt.Println("Starting GitHub MCP Server...")
	
	// Check critical environment variables
	token := os.Getenv("GITHUB_PERSONAL_ACCESS_TOKEN")
	if token == "" {
		fmt.Println("ERROR: GITHUB_PERSONAL_ACCESS_TOKEN environment variable not set!")
		fmt.Println("Please set a valid GitHub token with appropriate permissions.")
		os.Exit(1)
	}
	
	// Log that we have a token (but don't print it fully)
	fmt.Printf("GitHub token found, starts with: %s... (length: %d)\n", token[:5], len(token))
	
	// Add more startup diagnostic info
	fmt.Println("Testing connectivity to GitHub API...")
	
	// Create a test client
	ctx := context.Background()
	sts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, sts)
	gqlClient := githubv4.NewClient(httpClient)
	
	// Try a simple GraphQL query
	var query struct {
		Viewer struct {
			Login string
		}
	}
	
	err := gqlClient.Query(ctx, &query, nil)
	if err != nil {
		fmt.Printf("ERROR: Could not connect to GitHub API: %v\n", err)
		fmt.Println("Please check your token permissions and network connectivity.")
	} else {
		fmt.Printf("Successfully authenticated to GitHub API as: %s\n", query.Viewer.Login)
	}
	
	// Continue with server startup
	fmt.Println("Initializing MCP server with GitHub tools...")
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
