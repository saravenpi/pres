package catalyst

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/saravenpi/catalyst/internal/builder"
	"github.com/saravenpi/catalyst/internal/server"
	"github.com/spf13/cobra"
)

func Execute(version string) {
	rootCmd := buildRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCmd(version string) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "catalyst",
		Short:   "Markdown-to-slides CLI tool",
		Long:    "Catalyst turns a folder of markdown files into a self-contained HTML slideshow.",
		Version: version,
	}
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")

	rootCmd.AddCommand(
		serveCmd(),
		buildCmd(),
		newCmd(),
	)

	return rootCmd
}

func serveCmd() *cobra.Command {
	var portFlag int
	cmd := &cobra.Command{
		Use:   "serve <dir>",
		Short: "Parse, render, and serve a presentation with live reload",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return server.Serve(server.Options{
				Dir:  args[0],
				Port: portFlag,
			})
		},
	}
	cmd.Flags().IntVarP(&portFlag, "port", "p", 3000, "port to serve on")
	return cmd
}

func buildCmd() *cobra.Command {
	var outFlag string
	cmd := &cobra.Command{
		Use:   "build <dir>",
		Short: "Compile a presentation to a self-contained output file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return builder.Build(builder.Options{
				Dir:    args[0],
				Out: outFlag,
			})
		},
	}
	cmd.Flags().StringVarP(&outFlag, "output", "o", "dist", "output file")
	return cmd
}

func newCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "new <name>",
		Short: "Scaffold a new presentation folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if _, err := os.Stat(name); err == nil {
				return fmt.Errorf("directory %q already exists", name)
			}
			if err := os.Mkdir(name, 0755); err != nil {
				return err
			}

			presPath := filepath.Join(name, "pres.md")
			starter := fmt.Sprintf(
				"# %s\n\nWelcome to your Catalyst presentation!\n\n---\n\n"+
					"## Slide 2\n\nYour content here.\n\n---\n\n"+
					"## Slide 3\n\nMore content.\n",
				name,
			)
			if err := os.WriteFile(presPath, []byte(starter), 0644); err != nil {
				_ = os.RemoveAll(name)
				return fmt.Errorf("writing pres.md: %w", err)
			}

			fmt.Printf("✓ Created %s/pres.md\n", name)
			return nil
		},
	}
}