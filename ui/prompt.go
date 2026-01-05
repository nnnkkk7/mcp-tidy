package ui

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nnnkkk7/mcp-tidy/types"
)

// ConfirmPrompt asks the user for confirmation.
// Returns true for "y" or "yes", false otherwise.
func ConfirmPrompt(prompt string, defaultYes bool) bool {
	return ConfirmPromptWithReader(os.Stdin, os.Stdout, prompt, defaultYes)
}

// ConfirmPromptWithReader asks for confirmation with custom reader/writer (for testing).
func ConfirmPromptWithReader(r io.Reader, w io.Writer, prompt string, defaultYes bool) bool {
	hint := "[y/N]"
	if defaultYes {
		hint = "[Y/n]"
	}

	fmt.Fprintf(w, "%s %s ", prompt, hint)

	reader := bufio.NewReader(r)
	input, err := reader.ReadString('\n')
	if err != nil {
		return defaultYes
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" {
		return defaultYes
	}

	return input == "y" || input == "yes"
}

// SelectServersPrompt displays servers and lets user select which to remove.
// Returns the indices of selected servers.
func SelectServersPrompt(servers []types.MCPServer, stats map[string]types.ServerStats) []int {
	return SelectServersPromptWithReader(os.Stdin, os.Stdout, servers, stats)
}

// SelectServersPromptWithReader is testable version with custom reader/writer.
func SelectServersPromptWithReader(r io.Reader, w io.Writer, servers []types.MCPServer, stats map[string]types.ServerStats) []int {
	if len(servers) == 0 {
		return nil
	}

	fmt.Fprintln(w, "\n? Select servers to remove (enter numbers separated by spaces, or 'all'):")

	for i := range servers {
		stat, ok := stats[servers[i].Name]
		info := "(no usage data)"
		if ok {
			if stat.Calls == 0 {
				info = warningColor.Sprint("(0 calls, never used) ⚠️ unused")
			} else {
				info = fmt.Sprintf("(%d calls, %s)", stat.Calls, stat.LastUsedString())
			}
		}

		fmt.Fprintf(w, "  [%d] %s %s\n", i+1, servers[i].Name, info)
	}

	fmt.Fprint(w, "\nEnter selection: ")

	reader := bufio.NewReader(r)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil
	}

	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return nil
	}

	if input == "all" {
		result := make([]int, len(servers))
		for i := range servers {
			result[i] = i
		}
		return result
	}

	var selected []int
	for _, part := range strings.Fields(input) {
		var idx int
		if _, err := fmt.Sscanf(part, "%d", &idx); err == nil {
			if idx >= 1 && idx <= len(servers) {
				selected = append(selected, idx-1)
			}
		}
	}

	return selected
}
