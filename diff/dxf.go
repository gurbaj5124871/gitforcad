package diff

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// DXF entity representation
type dxfEntity struct {
	Type   string
	Layer  string
	Handle string
	Data   map[int]string // group code → value
}

// DiffDXF computes diff between two DXF files.
func DiffDXF(filePath string, oldContent, newContent []byte) *DiffResult {
	result := &DiffResult{
		FilePath: filePath,
		FileType: "dxf",
		IsBinary: false,
		IsCAD:    true,
	}

	var oldEntities, newEntities []dxfEntity

	if oldContent != nil {
		oldEntities = parseDXFEntities(oldContent)
	}
	if newContent != nil {
		newEntities = parseDXFEntities(newContent)
	}

	// New file
	if oldContent == nil {
		counts := entityCounts(newEntities)
		result.Summary = fmt.Sprintf("new DXF: %s (%d entities)", filePath, len(newEntities))
		for eType, count := range counts {
			result.Lines = append(result.Lines, DiffLine{Type: "add",
				Content: fmt.Sprintf("%s: %d entities", eType, count)})
		}
		result.Stats.Additions = len(newEntities)
		return result
	}

	// Deleted file
	if newContent == nil {
		result.Summary = fmt.Sprintf("deleted DXF: %s (%d entities)", filePath, len(oldEntities))
		counts := entityCounts(oldEntities)
		for eType, count := range counts {
			result.Lines = append(result.Lines, DiffLine{Type: "del",
				Content: fmt.Sprintf("%s: %d entities", eType, count)})
		}
		result.Stats.Deletions = len(oldEntities)
		return result
	}

	// Compare entities
	oldCounts := entityCounts(oldEntities)
	newCounts := entityCounts(newEntities)

	// Collect all entity types
	allTypes := make(map[string]bool)
	for t := range oldCounts {
		allTypes[t] = true
	}
	for t := range newCounts {
		allTypes[t] = true
	}

	var added, removed int
	for eType := range allTypes {
		oldCount := oldCounts[eType]
		newCount := newCounts[eType]

		if oldCount == newCount {
			result.Lines = append(result.Lines, DiffLine{Type: "ctx",
				Content: fmt.Sprintf("%s: %d (unchanged)", eType, oldCount)})
		} else {
			if oldCount > 0 {
				result.Lines = append(result.Lines, DiffLine{Type: "del",
					Content: fmt.Sprintf("%s: %d", eType, oldCount)})
				removed += oldCount
			}
			if newCount > 0 {
				result.Lines = append(result.Lines, DiffLine{Type: "add",
					Content: fmt.Sprintf("%s: %d", eType, newCount)})
				added += newCount
			}
		}
	}

	// Compare layers
	oldLayers := entityLayers(oldEntities)
	newLayers := entityLayers(newEntities)

	allLayers := make(map[string]bool)
	for l := range oldLayers {
		allLayers[l] = true
	}
	for l := range newLayers {
		allLayers[l] = true
	}

	if len(allLayers) > 0 {
		result.Lines = append(result.Lines, DiffLine{Type: "ctx", Content: "--- Layers ---"})
		for layer := range allLayers {
			_, inOld := oldLayers[layer]
			_, inNew := newLayers[layer]

			if inOld && !inNew {
				result.Lines = append(result.Lines, DiffLine{Type: "del",
					Content: fmt.Sprintf("Layer: %s (%d entities)", layer, oldLayers[layer])})
			} else if !inOld && inNew {
				result.Lines = append(result.Lines, DiffLine{Type: "add",
					Content: fmt.Sprintf("Layer: %s (%d entities)", layer, newLayers[layer])})
			} else if oldLayers[layer] != newLayers[layer] {
				result.Lines = append(result.Lines, DiffLine{Type: "del",
					Content: fmt.Sprintf("Layer: %s (%d entities)", layer, oldLayers[layer])})
				result.Lines = append(result.Lines, DiffLine{Type: "add",
					Content: fmt.Sprintf("Layer: %s (%d entities)", layer, newLayers[layer])})
			}
		}
	}

	result.Summary = fmt.Sprintf("DXF: %d → %d entities (+%d -%d)",
		len(oldEntities), len(newEntities), added, removed)
	result.Stats.Additions = added
	result.Stats.Deletions = removed

	return result
}

// parseDXFEntities extracts entities from the ENTITIES section of a DXF file.
func parseDXFEntities(data []byte) []dxfEntity {
	var entities []dxfEntity
	scanner := bufio.NewScanner(bytes.NewReader(data))

	inEntities := false
	var current *dxfEntity

	for scanner.Scan() {
		codeLine := strings.TrimSpace(scanner.Text())
		if !scanner.Scan() {
			break
		}
		valueLine := strings.TrimSpace(scanner.Text())

		code := 0
		fmt.Sscanf(codeLine, "%d", &code)

		// Track section
		if code == 2 && valueLine == "ENTITIES" {
			inEntities = true
			continue
		}
		if code == 0 && valueLine == "ENDSEC" && inEntities {
			if current != nil {
				entities = append(entities, *current)
			}
			break
		}

		if !inEntities {
			continue
		}

		if code == 0 {
			// New entity
			if current != nil {
				entities = append(entities, *current)
			}
			current = &dxfEntity{
				Type: valueLine,
				Data: make(map[int]string),
			}
		} else if current != nil {
			current.Data[code] = valueLine
			switch code {
			case 5:
				current.Handle = valueLine
			case 8:
				current.Layer = valueLine
			}
		}
	}

	return entities
}

func entityCounts(entities []dxfEntity) map[string]int {
	counts := make(map[string]int)
	for _, e := range entities {
		counts[e.Type]++
	}
	return counts
}

func entityLayers(entities []dxfEntity) map[string]int {
	layers := make(map[string]int)
	for _, e := range entities {
		if e.Layer != "" {
			layers[e.Layer]++
		}
	}
	return layers
}
