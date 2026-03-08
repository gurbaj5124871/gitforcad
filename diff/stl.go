package diff

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// STL Triangle representation
type stlTriangle struct {
	Normal   [3]float32
	Vertices [3][3]float32
}

// DiffSTL computes diff between two STL files.
func DiffSTL(filePath string, oldContent, newContent []byte) *DiffResult {
	result := &DiffResult{
		FilePath: filePath,
		FileType: "stl",
		IsBinary: true,
		IsCAD:    true,
	}

	var oldTris, newTris []stlTriangle
	var oldErr, newErr error

	if oldContent != nil {
		oldTris, oldErr = parseSTL(oldContent)
	}
	if newContent != nil {
		newTris, newErr = parseSTL(newContent)
	}

	// Handle parse errors — fall back to binary diff
	if oldErr != nil || newErr != nil {
		return DiffBinary(filePath, oldContent, newContent)
	}

	// Handle new/deleted file
	if oldContent == nil {
		result.Summary = fmt.Sprintf("new STL file: %s (%d triangles)", filePath, len(newTris))
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("STL mesh with %d triangles", len(newTris))})
		bb := boundingBox(newTris)
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Bounding box: (%.2f, %.2f, %.2f) → (%.2f, %.2f, %.2f)",
				bb[0], bb[1], bb[2], bb[3], bb[4], bb[5])})
		result.Stats.Additions = len(newTris)
		return result
	}

	if newContent == nil {
		result.Summary = fmt.Sprintf("deleted STL: %s (%d triangles removed)", filePath, len(oldTris))
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("STL mesh with %d triangles", len(oldTris))})
		result.Stats.Deletions = len(oldTris)
		return result
	}

	// Compare meshes
	oldSet := triangleSet(oldTris)
	newSet := triangleSet(newTris)

	var added, removed int
	for key := range newSet {
		if _, exists := oldSet[key]; !exists {
			added++
		}
	}
	for key := range oldSet {
		if _, exists := newSet[key]; !exists {
			removed++
		}
	}

	// Bounding boxes
	oldBB := boundingBox(oldTris)
	newBB := boundingBox(newTris)

	// Surface area approximation
	oldArea := totalSurfaceArea(oldTris)
	newArea := totalSurfaceArea(newTris)

	result.Summary = fmt.Sprintf("STL: %d triangles → %d triangles (+%d -%d)",
		len(oldTris), len(newTris), added, removed)

	// Summary lines
	if len(oldTris) != len(newTris) {
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Triangles: %d", len(oldTris))})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Triangles: %d", len(newTris))})
	} else {
		result.Lines = append(result.Lines, DiffLine{Type: "ctx",
			Content: fmt.Sprintf("Triangles: %d (unchanged count)", len(oldTris))})
	}

	if removed > 0 {
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Removed %d triangles from mesh", removed)})
	}
	if added > 0 {
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Added %d triangles to mesh", added)})
	}

	// Bounding box changes
	if oldBB != newBB {
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Bounding box: (%.2f, %.2f, %.2f) → (%.2f, %.2f, %.2f)",
				oldBB[0], oldBB[1], oldBB[2], oldBB[3], oldBB[4], oldBB[5])})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Bounding box: (%.2f, %.2f, %.2f) → (%.2f, %.2f, %.2f)",
				newBB[0], newBB[1], newBB[2], newBB[3], newBB[4], newBB[5])})
	}

	// Surface area
	if math.Abs(float64(oldArea-newArea)) > 0.01 {
		result.Lines = append(result.Lines, DiffLine{Type: "del",
			Content: fmt.Sprintf("Surface area: %.2f", oldArea)})
		result.Lines = append(result.Lines, DiffLine{Type: "add",
			Content: fmt.Sprintf("Surface area: %.2f", newArea)})
	}

	result.Stats.Additions = added
	result.Stats.Deletions = removed

	return result
}

// parseSTL parses both ASCII and binary STL files.
func parseSTL(data []byte) ([]stlTriangle, error) {
	// Detect ASCII vs binary: ASCII starts with "solid"
	if len(data) > 5 && strings.HasPrefix(strings.TrimSpace(string(data[:80])), "solid") {
		// But binary files can also start with "solid" in the header
		// Check if it looks like valid ASCII STL
		if bytes.Contains(data[:min(len(data), 1000)], []byte("facet normal")) {
			return parseASCIISTL(data)
		}
	}
	return parseBinarySTL(data)
}

func parseBinarySTL(data []byte) ([]stlTriangle, error) {
	if len(data) < 84 {
		return nil, fmt.Errorf("binary STL too short")
	}

	// Skip 80-byte header
	numTriangles := binary.LittleEndian.Uint32(data[80:84])
	expected := 84 + int(numTriangles)*50
	if len(data) < expected {
		return nil, fmt.Errorf("binary STL truncated: expected %d bytes, got %d", expected, len(data))
	}

	triangles := make([]stlTriangle, numTriangles)
	offset := 84

	for i := uint32(0); i < numTriangles; i++ {
		tri := &triangles[i]

		// Normal (3 floats)
		for j := 0; j < 3; j++ {
			tri.Normal[j] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
			offset += 4
		}

		// 3 vertices, each 3 floats
		for v := 0; v < 3; v++ {
			for j := 0; j < 3; j++ {
				tri.Vertices[v][j] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset:]))
				offset += 4
			}
		}

		// Skip attribute byte count (2 bytes)
		offset += 2
	}

	return triangles, nil
}

func parseASCIISTL(data []byte) ([]stlTriangle, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))
	var triangles []stlTriangle
	var current stlTriangle
	vertexIdx := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "facet normal") {
			fmt.Sscanf(line, "facet normal %f %f %f",
				&current.Normal[0], &current.Normal[1], &current.Normal[2])
			vertexIdx = 0
		} else if strings.HasPrefix(line, "vertex") {
			if vertexIdx < 3 {
				fmt.Sscanf(line, "vertex %f %f %f",
					&current.Vertices[vertexIdx][0],
					&current.Vertices[vertexIdx][1],
					&current.Vertices[vertexIdx][2])
				vertexIdx++
			}
		} else if strings.HasPrefix(line, "endfacet") {
			triangles = append(triangles, current)
			current = stlTriangle{}
		}
	}

	return triangles, nil
}

func triangleKey(t stlTriangle) string {
	return fmt.Sprintf("%.4f,%.4f,%.4f|%.4f,%.4f,%.4f|%.4f,%.4f,%.4f",
		t.Vertices[0][0], t.Vertices[0][1], t.Vertices[0][2],
		t.Vertices[1][0], t.Vertices[1][1], t.Vertices[1][2],
		t.Vertices[2][0], t.Vertices[2][1], t.Vertices[2][2])
}

func triangleSet(tris []stlTriangle) map[string]bool {
	set := make(map[string]bool)
	for _, t := range tris {
		set[triangleKey(t)] = true
	}
	return set
}

func boundingBox(tris []stlTriangle) [6]float32 {
	if len(tris) == 0 {
		return [6]float32{}
	}

	minX, minY, minZ := float32(math.MaxFloat32), float32(math.MaxFloat32), float32(math.MaxFloat32)
	maxX, maxY, maxZ := float32(-math.MaxFloat32), float32(-math.MaxFloat32), float32(-math.MaxFloat32)

	for _, t := range tris {
		for _, v := range t.Vertices {
			if v[0] < minX {
				minX = v[0]
			}
			if v[1] < minY {
				minY = v[1]
			}
			if v[2] < minZ {
				minZ = v[2]
			}
			if v[0] > maxX {
				maxX = v[0]
			}
			if v[1] > maxY {
				maxY = v[1]
			}
			if v[2] > maxZ {
				maxZ = v[2]
			}
		}
	}

	return [6]float32{minX, minY, minZ, maxX, maxY, maxZ}
}

func totalSurfaceArea(tris []stlTriangle) float32 {
	var total float32
	for _, t := range tris {
		// Cross product of two edges
		ax := t.Vertices[1][0] - t.Vertices[0][0]
		ay := t.Vertices[1][1] - t.Vertices[0][1]
		az := t.Vertices[1][2] - t.Vertices[0][2]
		bx := t.Vertices[2][0] - t.Vertices[0][0]
		by := t.Vertices[2][1] - t.Vertices[0][1]
		bz := t.Vertices[2][2] - t.Vertices[0][2]

		cx := ay*bz - az*by
		cy := az*bx - ax*bz
		cz := ax*by - ay*bx

		total += float32(math.Sqrt(float64(cx*cx+cy*cy+cz*cz))) / 2
	}
	return total
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
