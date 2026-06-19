package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
				apiURL = "http://localhost:8090"
			}
		},
	}

	root.AddCommand(tasksCmd(), notesCmd(), transactionsCmd())

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
