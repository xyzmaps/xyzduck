package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/spf13/cobra"
	"org.xyzmaps.xyzduck/src/database"
)

var initCmd = &cobra.Command{
	Use:   "init [filename]",
	Short: "Initialize a DuckDB database with spatial extension",
	Long: `Create a new DuckDB database file or open an existing one and ensure
the spatial extension is installed and loaded. If no filename is provided,
an interactive prompt will ask for the database name.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	var filename string
	var err error

	// Check if filename was provided as argument
	if len(args) > 0 {
		filename = args[0]
	} else {
		// No argument provided, launch TUI to prompt for filename
		filename, err = promptForFilename()
		if err != nil {
			return err
		}
	}

	// Validate filename
	if strings.TrimSpace(filename) == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Ensure .duckdb extension
	filename = database.EnsureDuckDBExtension(filename)

	// Check if file exists
	exists := database.FileExists(filename)
	if exists {
		fmt.Printf("Opening existing database: %s\n", filename)
	} else {
		fmt.Printf("Creating new database: %s\n", filename)
	}

	// Create or open the database
	if err := database.CreateOrOpenDatabase(filename); err != nil {
		return fmt.Errorf("failed to create/open database: %w", err)
	}

	// Initialize spatial extension
	fmt.Println("Installing spatial extension...")
	if err := database.InitSpatialExtension(filename); err != nil {
		return fmt.Errorf("failed to initialize spatial extension: %w", err)
	}

	fmt.Printf("\n✓ Database ready with spatial extension at: %s\n", filename)
	return nil
}

// TUI Model for filename input
type filenameModel struct {
	textInput textinput.Model
	err       error
	submitted bool
	cancelled bool
}

func initialModel() filenameModel {
	ti := textinput.New()
	ti.Placeholder = "mydata"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 50

	return filenameModel{
		textInput: ti,
		err:       nil,
		submitted: false,
		cancelled: false,
	}
}

func (m filenameModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m filenameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Validate input
			value := strings.TrimSpace(m.textInput.Value())
			if value == "" {
				m.err = fmt.Errorf("filename cannot be empty")
				return m, nil
			}
			m.submitted = true
			m.err = nil
			return m, tea.Quit

		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m filenameModel) View() string {
	if m.submitted {
		return ""
	}

	s := "\nWhat would you like to name your database?\n\n"
	s += m.textInput.View() + "\n\n"

	if m.err != nil {
		s += fmt.Sprintf("Error: %s\n\n", m.err)
	}

	s += "(esc to cancel)\n"

	return s
}

// promptForFilename launches the Bubble Tea TUI to get filename from user
func promptForFilename() (string, error) {
	p := tea.NewProgram(initialModel())
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running prompt: %w", err)
	}

	m := finalModel.(filenameModel)

	if m.cancelled {
		return "", fmt.Errorf("cancelled by user")
	}

	if m.err != nil {
		return "", m.err
	}

	return strings.TrimSpace(m.textInput.Value()), nil
}
