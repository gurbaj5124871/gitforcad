package diff

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

// HasLibreDWG checks if the dwgread executable is available in the system PATH.
func HasLibreDWG() bool {
	_, err := exec.LookPath("dwgread")
	return err == nil
}

// DiffDWG converts DWG files to DXF using LibreDWG's dwgread, then diffs the DXF.
func DiffDWG(filePath string, oldContent, newContent []byte) *DiffResult {
	if !HasLibreDWG() {
		return &DiffResult{
			IsCAD:    true,
			FileType: "dwg",
			FilePath: filePath,
			Summary:  "Error: LibreDWG (dwgread) not installed in PATH",
			Lines:    []DiffLine{{Type: "ctx", Content: "DWG diffs require 'dwgread' from the GNU LibreDWG package."}},
			Stats:    DiffStats{Additions: 1, Deletions: 1},
		}
	}

	// Create temp files for the DWG binary content
	oldTmp, err := writeTempDWG(oldContent)
	if err != nil {
		result := DiffBinary(filePath, oldContent, newContent)
		result.Summary = fmt.Sprintf("failed to save old dwg: %v", err)
		return result
	}
	defer os.Remove(oldTmp)

	newTmp, err := writeTempDWG(newContent)
	if err != nil {
		result := DiffBinary(filePath, oldContent, newContent)
		result.Summary = fmt.Sprintf("failed to save new dwg: %v", err)
		return result
	}
	defer os.Remove(newTmp)

	// Run dwgread -O DXF <file>
	oldDXF, err := runDWGRead(oldTmp)
	if err != nil {
		// If both fail parsing, fallback to binary
		if _, err2 := runDWGRead(newTmp); err2 != nil {
			return DiffBinary(filePath, oldContent, newContent)
		}
	}

	newDXF, _ := runDWGRead(newTmp)

	var oldDXFBytes, newDXFBytes []byte
	if oldDXF != "" {
		oldDXFBytes = []byte(oldDXF)
	}
	if newDXF != "" {
		newDXFBytes = []byte(newDXF)
	}

	// Delegate to our native DXF diff engine
	result := DiffDXF(filePath, oldDXFBytes, newDXFBytes)

	// Rewrite the format and headline to reflect it's DWG under the hood
	result.FileType = "dwg"
	if result.Summary != "" && len(result.Summary) > 4 {
		result.Summary = "DWG (via DXF): " + result.Summary[5:] // replacing "DXF: "
	}

	return result
}

func writeTempDWG(content []byte) (string, error) {
	if len(content) == 0 {
		return "", nil // empty file (e.g., added/deleted file)
	}

	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:10])
	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("gitforcad_%s.dwg", hashStr))

	if err := ioutil.WriteFile(tmpPath, content, 0644); err != nil {
		return "", err
	}
	return tmpPath, nil
}

func runDWGRead(path string) (string, error) {
	if path == "" {
		return "", nil // empty input (e.g. file didn't exist in old commit)
	}

	// dwgread -O DXF file.dwg
	cmd := exec.Command("dwgread", "-O", "DXF", path)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("dwgread failed: %v\nOutput: %s", err, string(out))
	}
	return string(out), nil
}
