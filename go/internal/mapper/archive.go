package mapper

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

// MaxArchiveSize matches the existing upload API's 1 MiB total-size limit.
// It protects the standalone CLI now and is the limit future HTTP handlers
// should enforce before decrypting or decoding a payload.
const MaxArchiveSize int64 = 1 * 1024 * 1024

// Drop is the language-neutral form used by the renderer and summary code.
type Drop struct {
	SiteID     int
	ResourceID int
	PositionX  float64
	PositionZ  float64
}

// ParseArchive extracts valid harvesting drops from a decrypted MsgPack payload.
// The game payload has changed field encodings across releases, so values are
// decoded dynamically and coerced only after explicit range/finite checks.
// Malformed optional records are skipped instead of aborting an otherwise valid
// archive, matching the Python pipeline's tolerant extraction behavior.
func ParseArchive(plain []byte) ([]Drop, error) {
	var payload map[string]any
	if err := msgpack.Unmarshal(plain, &payload); err != nil {
		return nil, fmt.Errorf("decode MsgPack archive: %w", err)
	}

	updated, ok := asStringMap(payload["updatedResources"])
	if !ok {
		return []Drop{}, nil
	}
	harvestMaps, ok := asSlice(updated["userMysekaiHarvestMaps"])
	if !ok {
		return []Drop{}, nil
	}

	drops := make([]Drop, 0)
	for _, siteValue := range harvestMaps {
		site, ok := asStringMap(siteValue)
		if !ok {
			continue
		}
		siteID, ok := asPositiveInt(site["mysekaiSiteId"])
		if !ok {
			continue
		}
		rawDrops, ok := asSlice(site["userMysekaiSiteHarvestResourceDrops"])
		if !ok {
			continue
		}
		for _, rawDropValue := range rawDrops {
			rawDrop, ok := asStringMap(rawDropValue)
			if !ok {
				continue
			}
			resourceID, resourceOK := asPositiveInt(rawDrop["resourceId"])
			x, xOK := asFiniteFloat(rawDrop["positionX"])
			z, zOK := asFiniteFloat(rawDrop["positionZ"])
			if !resourceOK || !xOK || !zOK {
				continue
			}
			drops = append(drops, Drop{
				SiteID:     siteID,
				ResourceID: resourceID,
				PositionX:  x,
				PositionZ:  z,
			})
		}
	}
	return drops, nil
}

func asStringMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			keyString, ok := key.(string)
			if !ok {
				continue
			}
			result[keyString] = item
		}
		return result, true
	default:
		return nil, false
	}
}

func asSlice(value any) ([]any, bool) {
	values, ok := value.([]any)
	return values, ok
}

func asPositiveInt(value any) (int, bool) {
	var number float64
	switch typed := value.(type) {
	case int:
		return positiveIntFromInt64(int64(typed))
	case int8:
		return positiveIntFromInt64(int64(typed))
	case int16:
		return positiveIntFromInt64(int64(typed))
	case int32:
		return positiveIntFromInt64(int64(typed))
	case int64:
		return positiveIntFromInt64(typed)
	case uint:
		return positiveIntFromUint64(uint64(typed))
	case uint8:
		return positiveIntFromUint64(uint64(typed))
	case uint16:
		return positiveIntFromUint64(uint64(typed))
	case uint32:
		return positiveIntFromUint64(uint64(typed))
	case uint64:
		return positiveIntFromUint64(typed)
	case float32:
		number = float64(typed)
	case float64:
		number = typed
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, false
		}
		number = parsed
	default:
		return 0, false
	}
	if math.IsNaN(number) || math.IsInf(number, 0) || number <= 0 || number > float64(maxInt()) {
		return 0, false
	}
	return int(number), true
}

func positiveIntFromInt64(value int64) (int, bool) {
	if value <= 0 || value > int64(maxInt()) {
		return 0, false
	}
	return int(value), true
}

func positiveIntFromUint64(value uint64) (int, bool) {
	if value == 0 || value > uint64(maxInt()) {
		return 0, false
	}
	return int(value), true
}

func maxInt() int {
	return int(^uint(0) >> 1)
}

func asFiniteFloat(value any) (float64, bool) {
	var number float64
	switch typed := value.(type) {
	case int:
		number = float64(typed)
	case int8:
		number = float64(typed)
	case int16:
		number = float64(typed)
	case int32:
		number = float64(typed)
	case int64:
		number = float64(typed)
	case uint:
		number = float64(typed)
	case uint8:
		number = float64(typed)
	case uint16:
		number = float64(typed)
	case uint32:
		number = float64(typed)
	case uint64:
		number = float64(typed)
	case float32:
		number = float64(typed)
	case float64:
		number = typed
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err != nil {
			return 0, false
		}
		number = parsed
	default:
		return 0, false
	}
	if math.IsNaN(number) || math.IsInf(number, 0) {
		return 0, false
	}
	return number, true
}

// ReadDrops decrypts and parses a save file. It never writes plaintext to disk
// and reads at most MaxArchiveSize+1 bytes to reject oversized input safely.
func ReadDrops(path string) ([]Drop, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()

	raw, err := io.ReadAll(io.LimitReader(file, MaxArchiveSize+1))
	if err != nil {
		return nil, fmt.Errorf("read archive: %w", err)
	}
	if int64(len(raw)) > MaxArchiveSize {
		return nil, fmt.Errorf("archive exceeds %d-byte limit", MaxArchiveSize)
	}
	plain, err := DecryptArchive(raw)
	if err != nil {
		return nil, err
	}
	return ParseArchive(plain)
}

// Rotate matches the existing Python coordinate normalization exactly.
func Rotate(x, z float64, siteID int) (float64, float64) {
	switch siteID {
	case 6:
		return z, -x
	case 5, 8:
		return -z, x
	case 7:
		return -x, -z
	default:
		return x, z
	}
}

// Summary contains no raw save payload or player identifier and is suitable for
// local validation output.
type Summary struct {
	TotalDrops        int         `json:"total_drops"`
	SiteCounts        map[int]int `json:"site_counts"`
	RareCounts        map[int]int `json:"rare_counts"`
	DistinctResources int         `json:"distinct_resources"`
}

func Summarize(drops []Drop) Summary {
	summary := Summary{
		TotalDrops: len(drops),
		SiteCounts: map[int]int{},
		RareCounts: map[int]int{5: 0, 12: 0, 20: 0, 24: 0},
	}
	resources := map[int]struct{}{}
	for _, drop := range drops {
		summary.SiteCounts[drop.SiteID]++
		resources[drop.ResourceID] = struct{}{}
		if _, rare := summary.RareCounts[drop.ResourceID]; rare {
			summary.RareCounts[drop.ResourceID]++
		}
	}
	summary.DistinctResources = len(resources)
	return summary
}

func sortedSiteIDs(drops []Drop) []int {
	seen := map[int]struct{}{}
	for _, drop := range drops {
		seen[drop.SiteID] = struct{}{}
	}
	ids := make([]int, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
