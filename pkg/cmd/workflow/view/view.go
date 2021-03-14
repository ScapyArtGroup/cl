package view

import (
	"errors"
	"net/http"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/internal/ghrepo"
	"github.com/cli/cli/pkg/cmdutil"
	"github.com/cli/cli/pkg/iostreams"
	"github.com/spf13/cobra"
)

type ViewOptions struct {
	HttpClient func() (*http.Client, error)
	IO         *iostreams.IOStreams
	BaseRepo   func() (ghrepo.Interface, error)

	WorkflowID string
	YAML       string

	Prompt bool
}

func NewCmdView(f *cmdutil.Factory, runF func(*ViewOptions) error) *cobra.Command {
	opts := &ViewOptions{
		IO:         f.IOStreams,
		HttpClient: f.HttpClient,
	}
	cmd := &cobra.Command{
		Use:    "view [<run-id>]",
		Short:  "View a summary of a workflow run",
		Args:   cobra.MaximumNArgs(1),
		Hidden: true,
		Example: heredoc.Doc(`
		  # Interactively select a workflow to view
		  $ gh workflow view

		  # View a specific workflow
		  $ gh workflow view 0451
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			// support `-R, --repo` override
			opts.BaseRepo = f.BaseRepo

			if len(args) > 0 {
				opts.WorkflowID = args[0]
			} else if !opts.IO.CanPrompt() {
				return &cmdutil.FlagError{Err: errors.New("workflow ID required when not running interactively")}
			} else {
				opts.Prompt = true
			}

			if runF != nil {
				return runF(opts)
			}
			return runView(opts)
		},
	}

	return cmd
}

func runView(opts *ViewOptions) error {
	// TODO
	return nil
}
