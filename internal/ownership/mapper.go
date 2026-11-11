package ownership

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type CodeOwnership struct {
	Team          string `json:"team"`
	Lead          string `json:"lead"`
	PrimaryAuthor string `json:"primary_author"`
}

type Mapper struct{}

func NewMapper() *Mapper {
	return &Mapper{}
}

// MapFindingOwnership determines team/author responsibility via CODEOWNERS or git blame
func (m *Mapper) MapFindingOwnership(repoPath, filePath string, line int) CodeOwnership {
	ownership := CodeOwnership{
		Team:          "Engineering",
		Lead:          "Unassigned",
		PrimaryAuthor: "Unknown",
	}

	// 1. Try reading CODEOWNERS file
	codeownersPath := filepath.Join(repoPath, ".github", "CODEOWNERS")
	if _, err := os.Stat(codeownersPath); err != nil {
		codeownersPath = filepath.Join(repoPath, "CODEOWNERS")
	}

	data, err := os.ReadFile(codeownersPath)
	if err == nil {
		lines := strings.Split(string(data), "\n")
		for _, l := range lines {
			l = strings.TrimSpace(l)
			if l == "" || strings.HasPrefix(l, "#") {
				continue
			}
			parts := strings.Fields(l)
			if len(parts) >= 2 {
				pattern := parts[0]
				owner := parts[1]
				if pattern == "*" || strings.Contains(filePath, strings.Trim(pattern, "*")) {
					ownership.Team = owner
					break
				}
			}
		}
	}

	// 2. Query git blame for exact author of line
	cmd := exec.Command("git", "blame", "-L", strings.Repeat("1,1", 0), "-e", filePath)
	if line > 0 {
		cmd = exec.Command("git", "blame", "-L", fmt.Sprintf("%d,+1", line), "-e", filePath)
	}
	cmd.Dir = repoPath
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err == nil {
		blameStr := out.String()
		if idx := strings.Index(blameStr, "("); idx != -1 {
			endIdx := strings.Index(blameStr[idx:], ")")
			if endIdx != -1 {
				authorPart := blameStr[idx+1 : idx+endIdx]
				fields := strings.Fields(authorPart)
				if len(fields) > 0 {
					ownership.PrimaryAuthor = fields[0]
				}
			}
		}
	}

	return ownership
}
