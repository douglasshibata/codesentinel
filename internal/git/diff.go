package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type ChangedFile struct {
	Path      string
	OldPath   string
	Status    string // "added", "modified", "deleted", "renamed"
	Additions int
	Deletions int
	Hunks     []DiffHunk
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Content  string
}

type DiffParser struct{}

func NewDiffParser() *DiffParser {
	return &DiffParser{}
}

// GetDiff executes `git diff <from>..<to>` or `git diff <from>` in repoPath
func (p *DiffParser) GetDiff(repoPath, fromBranch, toBranch string) (string, error) {
	var cmd *exec.Cmd
	if toBranch != "" && fromBranch != "" {
		cmd = exec.Command("git", "diff", fmt.Sprintf("%s..%s", fromBranch, toBranch))
	} else if fromBranch != "" {
		cmd = exec.Command("git", "diff", fromBranch)
	} else {
		cmd = exec.Command("git", "diff", "HEAD~1")
	}

	cmd.Dir = repoPath
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Fallback: git diff HEAD if HEAD~1 fails
		fallbackCmd := exec.Command("git", "diff", "HEAD")
		fallbackCmd.Dir = repoPath
		out.Reset()
		fallbackCmd.Stdout = &out
		if fallbackErr := fallbackCmd.Run(); fallbackErr == nil {
			return out.String(), nil
		}
		return "", fmt.Errorf("git diff failed: %v - %s", err, stderr.String())
	}

	return out.String(), nil
}

// ParseRawDiff parses unified diff text into structured ChangedFile slices
func (p *DiffParser) ParseRawDiff(rawDiff string) ([]ChangedFile, error) {
	var files []ChangedFile
	lines := strings.Split(rawDiff, "\n")

	var currentFile *ChangedFile

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if strings.HasPrefix(line, "diff --git ") {
			if currentFile != nil {
				files = append(files, *currentFile)
			}
			parts := strings.Fields(line)
			filePath := ""
			if len(parts) >= 4 {
				filePath = strings.TrimPrefix(parts[3], "b/")
			}
			currentFile = &ChangedFile{
				Path:   filePath,
				Status: "modified",
				Hunks:  []DiffHunk{},
			}
		} else if currentFile != nil {
			if strings.HasPrefix(line, "new file mode ") {
				currentFile.Status = "added"
			} else if strings.HasPrefix(line, "deleted file mode ") {
				currentFile.Status = "deleted"
			} else if strings.HasPrefix(line, "@@ ") {
				hunk := parseHunkHeader(line)
				currentFile.Hunks = append(currentFile.Hunks, hunk)
			} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				currentFile.Additions++
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				currentFile.Deletions++
			}
		}
	}

	if currentFile != nil {
		files = append(files, *currentFile)
	}

	return files, nil
}

func parseHunkHeader(header string) DiffHunk {
	// Header format: @@ -oldStart,oldLines +newStart,newLines @@
	hunk := DiffHunk{Content: header}
	parts := strings.Split(header, "@@")
	if len(parts) < 2 {
		return hunk
	}

	ranges := strings.Fields(strings.TrimSpace(parts[1]))
	if len(ranges) >= 2 {
		oldParts := strings.Split(strings.TrimPrefix(ranges[0], "-"), ",")
		if len(oldParts) >= 1 {
			hunk.OldStart, _ = strconv.Atoi(oldParts[0])
		}
		if len(oldParts) >= 2 {
			hunk.OldLines, _ = strconv.Atoi(oldParts[1])
		}

		newParts := strings.Split(strings.TrimPrefix(ranges[1], "+"), ",")
		if len(newParts) >= 1 {
			hunk.NewStart, _ = strconv.Atoi(newParts[0])
		}
		if len(newParts) >= 2 {
			hunk.NewLines, _ = strconv.Atoi(newParts[1])
		}
	}

	return hunk
}
