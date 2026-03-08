package diff

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// OBJ mesh data
type objMesh struct {
	Vertices  []string // Raw vertex lines for comparison
	Normals   []string
	TexCoords []string
	Faces     []string
	Groups    []string
	Materials []string
}

// DiffOBJ computes diff between two OBJ files.
func DiffOBJ(filePath string, oldContent, newContent []byte) *DiffResult {
	result := &DiffResult{
		FilePath: filePath,
		FileType: "obj",
		IsBinary: false,
		IsCAD:    true,
	}

	var oldMesh, newMesh *objMesh

	if oldContent != nil {
		oldMesh = parseOBJ(oldContent)
	}
	if newContent != nil {
		newMesh = parseOBJ(newContent)
	}

	// New file
	if oldContent == nil {
		result.Summary = fmt.Sprintf("new OBJ: %s (%d vertices, %d faces)",
			filePath, len(newMesh.Vertices), len(newMesh.Faces))
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Vertices: %d", len(newMesh.Vertices))})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Faces: %d", len(newMesh.Faces))})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Normals: %d", len(newMesh.Normals))})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Texture coordinates: %d", len(newMesh.TexCoords))})
		result.Stats.Additions = len(newMesh.Vertices) + len(newMesh.Faces)
		return result
	}

	// Deleted file
	if newContent == nil {
		result.Summary = fmt.Sprintf("deleted OBJ: %s (%d vertices, %d faces)",
			filePath, len(oldMesh.Vertices), len(oldMesh.Faces))
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Vertices: %d", len(oldMesh.Vertices))})
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Faces: %d", len(oldMesh.Faces))})
		result.Stats.Deletions = len(oldMesh.Vertices) + len(oldMesh.Faces)
		return result
	}

	// Compare meshes
	vertDiff := len(newMesh.Vertices) - len(oldMesh.Vertices)
	faceDiff := len(newMesh.Faces) - len(oldMesh.Faces)
	normDiff := len(newMesh.Normals) - len(oldMesh.Normals)

	// Vertex changes
	if vertDiff != 0 {
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Vertices: %d", len(oldMesh.Vertices))})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Vertices: %d (%+d)", len(newMesh.Vertices), vertDiff)})
	} else {
		// Check if vertices changed values
		changedVerts := countDifferences(oldMesh.Vertices, newMesh.Vertices)
		if changedVerts > 0 {
			result.Lines = append(result.Lines, DiffLine{Type: "ctx",
				Content: fmt.Sprintf("Vertices: %d (count unchanged, %d modified)", len(newMesh.Vertices), changedVerts)})
		} else {
			result.Lines = append(result.Lines, DiffLine{Type: "ctx",
				Content: fmt.Sprintf("Vertices: %d (unchanged)", len(newMesh.Vertices))})
		}
	}

	// Face changes
	if faceDiff != 0 {
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Faces: %d", len(oldMesh.Faces))})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Faces: %d (%+d)", len(newMesh.Faces), faceDiff)})
	} else {
		result.Lines = append(result.Lines, DiffLine{Type: "ctx",
			Content: fmt.Sprintf("Faces: %d (unchanged)", len(newMesh.Faces))})
	}

	// Normal changes
	if normDiff != 0 {
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Normals: %d", len(oldMesh.Normals))})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Normals: %d (%+d)", len(newMesh.Normals), normDiff)})
	}

	// Group changes
	oldGroups := stringSet(oldMesh.Groups)
	newGroups := stringSet(newMesh.Groups)
	for g := range newGroups {
		if _, ok := oldGroups[g]; !ok {
			result.Lines = append(result.Lines, DiffLine{Type: "add",
				Content: fmt.Sprintf("Group added: %s", g)})
		}
	}
	for g := range oldGroups {
		if _, ok := newGroups[g]; !ok {
			result.Lines = append(result.Lines, DiffLine{Type: "del",
				Content: fmt.Sprintf("Group removed: %s", g)})
		}
	}

	totalAdded := max(0, vertDiff) + max(0, faceDiff)
	totalRemoved := max(0, -vertDiff) + max(0, -faceDiff)

	result.Summary = fmt.Sprintf("OBJ: vertices %d→%d, faces %d→%d",
		len(oldMesh.Vertices), len(newMesh.Vertices),
		len(oldMesh.Faces), len(newMesh.Faces))
	result.Stats.Additions = totalAdded
	result.Stats.Deletions = totalRemoved

	return result
}

func parseOBJ(data []byte) *objMesh {
	mesh := &objMesh{}
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "v":
			mesh.Vertices = append(mesh.Vertices, line)
		case "vn":
			mesh.Normals = append(mesh.Normals, line)
		case "vt":
			mesh.TexCoords = append(mesh.TexCoords, line)
		case "f":
			mesh.Faces = append(mesh.Faces, line)
		case "g", "o":
			if len(parts) > 1 {
				mesh.Groups = append(mesh.Groups, parts[1])
			}
		case "usemtl":
			if len(parts) > 1 {
				mesh.Materials = append(mesh.Materials, parts[1])
			}
		}
	}

	return mesh
}

func countDifferences(a, b []string) int {
	count := 0
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	for i := 0; i < l; i++ {
		if a[i] != b[i] {
			count++
		}
	}
	return count
}

func stringSet(items []string) map[string]bool {
	s := make(map[string]bool)
	for _, item := range items {
		s[item] = true
	}
	return s
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
