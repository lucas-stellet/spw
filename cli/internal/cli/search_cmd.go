package cli

import (
	"fmt"
	"os"

	"github.com/lucas-stellet/oraculo/internal/store"
	"github.com/lucas-stellet/oraculo/internal/tools"
	"github.com/spf13/cobra"
)

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Full-text search across indexed specs",
		Long: `Search across all finalized spec documents using FTS5 full-text search.
Requires specs to have been indexed via "oraculo finalizar".`,
		Args: cobra.ExactArgs(1),
		Run:  runSearch,
	}

	cmd.Flags().String("spec", "", "Filter results to a specific spec")
	cmd.Flags().Int("limit", 5, "Maximum number of results")
	cmd.Flags().Bool("raw", false, "Output raw JSON")

	return cmd
}

func runSearch(cmd *cobra.Command, args []string) {
	raw, _ := cmd.Flags().GetBool("raw")
	specFilter, _ := cmd.Flags().GetString("spec")
	limit, _ := cmd.Flags().GetInt("limit")

	cwd := getCwd()
	query := args[0]

	ix, err := store.OpenIndex(cwd)
	if err != nil {
		if raw {
			tools.Fail("index not found: run 'oraculo finalizar <spec>' to index specs first", raw)
		}
		fmt.Fprintln(os.Stderr, "No search index found. Run 'oraculo finalizar <spec>' to index specs first.")
		os.Exit(1)
	}
	defer ix.Close()

	results, err := ix.Search(query, specFilter, limit)
	if err != nil {
		tools.Fail(fmt.Sprintf("search failed: %s", err), raw)
	}

	if raw {
		result := map[string]any{
			"ok":      true,
			"query":   query,
			"count":   len(results),
			"results": results,
		}
		tools.Output(result, "", true)
		return
	}

	if len(results) == 0 {
		fmt.Printf("No results found for %q\n", query)
		return
	}

	fmt.Printf("Search: %q\nFound %d results:\n\n", query, len(results))
	for i, r := range results {
		snippet := r.Snippet
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		fmt.Printf("  %d. [%s] %s\n", i+1, r.DocType, r.Spec)
		if r.Title != "" {
			fmt.Printf("     %s\n", r.Title)
		}
		if snippet != "" {
			fmt.Printf("     %s\n", snippet)
		}
		fmt.Printf("     Score: %.2f\n\n", r.Rank)
	}
}
