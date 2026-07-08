package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var apiURL string

func main() {
	root := &cobra.Command{
		Use:   "hube",
		Short: "hube personal hub CLI",
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			apiURL = os.Getenv("HUBE_URL")
			if apiURL == "" {
				// Matches the backend API's default port (see cmd/api/main.go
				// and frontend/vite.config.ts's dev proxy target).
				apiURL = "http://localhost:8080"
			}
		},
	}

	root.AddCommand(tasksCmd(), notesCmd(), transactionsCmd(), taskCmd(), noteCmd(), exportCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

// ── tasks ──────────────────────────────────────────────────────────────────

type task struct {
	ID       string     `json:"id"`
	Title    string     `json:"title"`
	Status   string     `json:"status"`
	Priority string     `json:"priority"`
	DueDate  *time.Time `json:"due_date"`
}

func tasksCmd() *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "List tasks",
		RunE: func(_ *cobra.Command, _ []string) error {
			url := apiURL + "/api/v1/tasks"
			if status != "" {
				url += "?status=" + status
			}

			var tasks []task
			if err := get(url, &tasks); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "STATUS\tPRIORITY\tTITLE\tDUE")
			for _, t := range tasks {
				due := "-"
				if t.DueDate != nil {
					due = t.DueDate.Format("2006-01-02")
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Status, t.Priority, t.Title, due)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status (todo, in_progress, done)")
	return cmd
}

// ── notes ──────────────────────────────────────────────────────────────────

type note struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func notesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "notes",
		Short: "List notes",
		RunE: func(_ *cobra.Command, _ []string) error {
			var notes []note
			if err := get(apiURL+"/api/v1/notes", &notes); err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TITLE\tPREVIEW")
			for _, n := range notes {
				preview := n.Content
				if len(preview) > 60 {
					preview = preview[:60] + "…"
				}
				fmt.Fprintf(w, "%s\t%s\n", n.Title, preview)
			}
			return w.Flush()
		},
	}
}

// ── note (single-resource CRUD) ─────────────────────────────────────────────

func noteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "note",
		Short: "Create, show, or update a single note",
	}
	cmd.AddCommand(noteCreateCmd(), noteShowCmd(), noteUpdateCmd())
	return cmd
}

func noteCreateCmd() *cobra.Command {
	var title, content string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a note",
		RunE: func(_ *cobra.Command, _ []string) error {
			body := map[string]any{"title": title, "content": content}
			var created note
			if err := postJSON(apiURL+"/api/v1/notes", body, &created); err != nil {
				return err
			}
			fmt.Printf("Created note %s: %s\n", created.ID, created.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "note title (required)")
	cmd.Flags().StringVar(&content, "content", "", "note content")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func noteShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a note by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var n note
			if err := get(apiURL+"/api/v1/notes/"+args[0], &n); err != nil {
				return err
			}
			fmt.Printf("ID:      %s\nTitle:   %s\nContent: %s\n", n.ID, n.Title, n.Content)
			return nil
		},
	}
}

func noteUpdateCmd() *cobra.Command {
	var title, status string

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a note's title and/or status",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			body := map[string]any{}
			if title != "" {
				body["title"] = title
			}
			if status != "" {
				body["status"] = status
			}
			var updated note
			if err := putJSON(apiURL+"/api/v1/notes/"+args[0], body, &updated); err != nil {
				return err
			}
			fmt.Printf("Updated note %s: %s\n", updated.ID, updated.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "new title")
	cmd.Flags().StringVar(&status, "status", "", "new status (draft, in_progress, published)")
	return cmd
}

// ── task (single-resource CRUD) ─────────────────────────────────────────────

func taskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Create a task or mark one as done",
	}
	cmd.AddCommand(taskCreateCmd(), taskDoneCmd())
	return cmd
}

func taskCreateCmd() *cobra.Command {
	var title, priority string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a task",
		RunE: func(_ *cobra.Command, _ []string) error {
			body := map[string]any{"title": title, "priority": priority}
			var created task
			if err := postJSON(apiURL+"/api/v1/tasks", body, &created); err != nil {
				return err
			}
			fmt.Printf("Created task %s: %s\n", created.ID, created.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "task title (required)")
	cmd.Flags().StringVar(&priority, "priority", "medium", "priority (low, medium, high)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func taskDoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "done <id>",
		Short: "Mark a task as done",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			var updated task
			if err := putJSON(apiURL+"/api/v1/tasks/"+args[0], map[string]any{"status": "done"}, &updated); err != nil {
				return err
			}
			fmt.Printf("Task %s marked done\n", updated.ID)
			return nil
		},
	}
}

// ── export ───────────────────────────────────────────────────────────────────

func exportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data as agent-friendly markdown",
	}
	cmd.AddCommand(exportNotesCmd())
	return cmd
}

func exportNotesCmd() *cobra.Command {
	var out string

	cmd := &cobra.Command{
		Use:   "notes",
		Short: "Export all notes as agent-friendly markdown (stdout, or one file per note with --out)",
		RunE: func(_ *cobra.Command, _ []string) error {
			var notes []note
			if err := get(apiURL+"/api/v1/notes", &notes); err != nil {
				return err
			}

			if out != "" {
				if err := os.MkdirAll(out, 0o755); err != nil {
					return fmt.Errorf("create output dir: %w", err)
				}
			}

			for _, n := range notes {
				md, err := getRaw(apiURL + "/api/v1/notes/" + n.ID + "/export")
				if err != nil {
					return fmt.Errorf("export note %s: %w", n.ID, err)
				}

				if out == "" {
					fmt.Println(md)
					continue
				}

				path := filepath.Join(out, slugify(n.Title)+".md")
				if err := os.WriteFile(path, []byte(md), 0o644); err != nil { //nolint:gosec
					return fmt.Errorf("write %s: %w", path, err)
				}
				fmt.Printf("Wrote %s\n", path)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&out, "out", "", "output directory (defaults to stdout)")
	return cmd
}

// slugify converts a note title into a filesystem-safe, lowercase, dash-separated
// filename stem (e.g. "My Note!" -> "my-note").
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		slug = "untitled"
	}
	return slug
}

// ── transactions ────────────────────────────────────────────────────────────

type financeSummary struct {
	Balance        float64 `json:"balance"`
	MonthIncome    float64 `json:"month_income"`
	MonthExpenses  float64 `json:"month_expenses"`
	RecentTransactions []struct {
		Amount      float64 `json:"amount"`
		Type        string  `json:"type"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Date        string  `json:"date"`
	} `json:"recent_transactions"`
}

func transactionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "transactions",
		Short: "Show finance summary and recent transactions",
		RunE: func(_ *cobra.Command, _ []string) error {
			var s financeSummary
			if err := get(apiURL+"/api/v1/finance/summary", &s); err != nil {
				return err
			}

			fmt.Printf("Balance:         $%.2f\n", s.Balance)
			fmt.Printf("Month income:    $%.2f\n", s.MonthIncome)
			fmt.Printf("Month expenses:  $%.2f\n\n", s.MonthExpenses)

			if len(s.RecentTransactions) == 0 {
				fmt.Println("No recent transactions.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "DATE\tTYPE\tCATEGORY\tAMOUNT\tDESCRIPTION")
			for _, t := range s.RecentTransactions {
				fmt.Fprintf(w, "%s\t%s\t%s\t$%.2f\t%s\n",
					t.Date, t.Type, t.Category, t.Amount, t.Description)
			}
			return w.Flush()
		},
	}
}

// ── helpers ─────────────────────────────────────────────────────────────────

func get(url string, dst any) error {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(dst)
}

// getRaw sends a GET request and returns the raw response body as a string,
// for endpoints that return plain text/markdown instead of JSON.
func getRaw(url string) (string, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}
	return string(b), nil
}

// postJSON sends a POST request with a JSON-encoded body and decodes the
// JSON response into dst (if non-nil).
func postJSON(url string, body, dst any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(b)) //nolint:gosec
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API returned %d", resp.StatusCode)
	}
	if dst == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}

// putJSON sends a PUT request with a JSON-encoded body and decodes the JSON
// response into dst (if non-nil).
func putJSON(url string, body, dst any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d", resp.StatusCode)
	}
	if dst == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(dst)
}
