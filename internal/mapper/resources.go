package mapper

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"image"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/webp"
)

// Resource is a named in-game resource icon decoded from assets/resourceId.csv.
type Resource struct {
	ID    int
	Name  string
	Image image.Image
}

// LoadResources accepts the repository resource CSV schema. A malformed icon
// is skipped rather than making a valid archive unusable.
func LoadResources(path string) (map[int]Resource, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open resource CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read resource CSV header: %w", err)
	}
	indices := map[string]int{}
	for i, name := range header {
		normalized := strings.TrimPrefix(strings.TrimSpace(name), "\ufeff")
		indices[normalized] = i
	}
	idIndex, idOK := indices["resourceId"]
	base64Index, base64OK := indices["base64"]
	nameIndex := -1
	for _, candidate := range []string{"name", "物品名", "itemName"} {
		if index, ok := indices[candidate]; ok {
			nameIndex = index
			break
		}
	}
	if !idOK || !base64OK {
		return nil, fmt.Errorf("resource CSV must contain resourceId and base64 columns")
	}

	resources := map[int]Resource{}
	for {
		record, readErr := reader.Read()
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return nil, fmt.Errorf("read resource CSV: %w", readErr)
		}
		if idIndex >= len(record) || base64Index >= len(record) {
			continue
		}
		id, parseErr := strconv.Atoi(strings.TrimSpace(record[idIndex]))
		if parseErr != nil || strings.TrimSpace(record[base64Index]) == "" {
			continue
		}
		raw, decodeErr := base64.StdEncoding.DecodeString(record[base64Index])
		if decodeErr != nil {
			continue
		}
		icon, decodeErr := webp.Decode(bytes.NewReader(raw))
		if decodeErr != nil {
			continue
		}
		name := strconv.Itoa(id)
		if nameIndex >= 0 && nameIndex < len(record) {
			name = record[nameIndex]
		}
		resources[id] = Resource{ID: id, Name: name, Image: icon}
	}
	if len(resources) == 0 {
		return nil, fmt.Errorf("resource CSV did not contain any decodable icons")
	}
	return resources, nil
}
