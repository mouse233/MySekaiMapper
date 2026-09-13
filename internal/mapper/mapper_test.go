package mapper

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/vmihailenco/msgpack/v5"
)

func encryptForTest(plain, key, iv []byte) []byte {
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	padded := append(append([]byte{}, plain...), bytes.Repeat([]byte{byte(pad)}, pad)...)
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	ciphertext := make([]byte, len(padded))
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(ciphertext, padded)
	return ciphertext
}

func TestDecryptWithKeyIVRoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")
	plain := []byte("hello mysekai")
	got, err := decryptWithKeyIV(encryptForTest(plain, key, iv), key, iv)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, plain) {
		t.Fatalf("got %q, want %q", got, plain)
	}
}

func TestDecryptRejectsInvalidPadding(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := []byte("fedcba9876543210")
	ciphertext := encryptForTest([]byte("padding validation"), key, iv)
	ciphertext[len(ciphertext)-1] ^= 0x01
	if _, err := decryptWithKeyIV(ciphertext, key, iv); err == nil {
		t.Fatal("expected invalid padding error")
	}
}

func TestParseArchiveAndSummary(t *testing.T) {
	payload := map[string]any{
		"updatedResources": map[string]any{
			"userMysekaiHarvestMaps": []any{
				map[string]any{
					"mysekaiSiteId": 5,
					"userMysekaiSiteHarvestResourceDrops": []any{
						map[string]any{"resourceId": 5, "positionX": 1.0, "positionZ": 2.0},
						map[string]any{"resourceId": 20, "positionX": 3.0, "positionZ": 4.0},
					},
				},
			},
		},
	}
	plain, err := msgpack.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	drops, err := ParseArchive(plain)
	if err != nil {
		t.Fatal(err)
	}
	if len(drops) != 2 || drops[0].SiteID != 5 || drops[1].ResourceID != 20 {
		t.Fatalf("unexpected drops: %#v", drops)
	}
	summary := Summarize(drops)
	if summary.TotalDrops != 2 || summary.RareCounts[5] != 1 || summary.RareCounts[20] != 1 {
		t.Fatalf("unexpected summary: %#v", summary)
	}
}

func TestParseArchiveCoercesSafeValuesAndSkipsInvalidRecords(t *testing.T) {
	payload := map[string]any{
		"updatedResources": map[string]any{
			"userMysekaiHarvestMaps": []any{
				map[string]any{
					"mysekaiSiteId": "5",
					"userMysekaiSiteHarvestResourceDrops": []any{
						map[string]any{"resourceId": "12", "positionX": int32(1), "positionZ": "2.5"},
						map[string]any{"resourceId": 20.9, "positionX": 3, "positionZ": 4},
						map[string]any{"resourceId": 24, "positionX": nil, "positionZ": 4},
						map[string]any{"resourceId": 5, "positionX": math.NaN(), "positionZ": 4},
					},
				},
				nil,
			},
		},
	}
	plain, err := msgpack.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	drops, err := ParseArchive(plain)
	if err != nil {
		t.Fatal(err)
	}
	want := []Drop{
		{SiteID: 5, ResourceID: 12, PositionX: 1, PositionZ: 2.5},
		{SiteID: 5, ResourceID: 20, PositionX: 3, PositionZ: 4},
	}
	if !reflect.DeepEqual(drops, want) {
		t.Fatalf("got %#v, want %#v", drops, want)
	}
}

func TestReadDropsRejectsOversizedInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "too-large.bin")
	if err := os.WriteFile(path, bytes.Repeat([]byte{0}, int(MaxArchiveSize)+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadDrops(path); err == nil {
		t.Fatal("expected oversized input to be rejected")
	}
}

func TestRotateMatchesSiteRules(t *testing.T) {
	cases := []struct {
		site         int
		x, z         float64
		wantX, wantZ float64
	}{
		{6, 1, 2, 2, -1},
		{5, 1, 2, -2, 1},
		{8, 1, 2, -2, 1},
		{7, 1, 2, -1, -2},
		{3, 1, 2, 1, 2},
	}
	for _, tc := range cases {
		gotX, gotZ := Rotate(tc.x, tc.z, tc.site)
		if gotX != tc.wantX || gotZ != tc.wantZ {
			t.Errorf("site %d: got (%v, %v), want (%v, %v)", tc.site, gotX, gotZ, tc.wantX, tc.wantZ)
		}
	}
}

func TestWriteRareResources(t *testing.T) {
	dir := t.TempDir()
	resources := map[int]Resource{
		5:  {ID: 5, Name: "A"},
		12: {ID: 12, Name: "B"},
		20: {ID: 20, Name: "C"},
		24: {ID: 24, Name: "D"},
	}
	drops := []Drop{{SiteID: 7, ResourceID: 12}, {SiteID: 5, ResourceID: 20}, {SiteID: 7, ResourceID: 20}}
	path, err := WriteRareResources(drops, resources, dir)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "稀有资源统计\n\nA × 0 \nB × 1 （烂漫花田）\nC × 2 （初始空地、烂漫花田）\nD × 0 "
	if string(got) != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("unexpected output path %q", path)
	}
}

func TestSummarySiteCounts(t *testing.T) {
	summary := Summarize([]Drop{{SiteID: 5, ResourceID: 1}, {SiteID: 5, ResourceID: 2}, {SiteID: 6, ResourceID: 1}})
	want := map[int]int{5: 2, 6: 1}
	if !reflect.DeepEqual(summary.SiteCounts, want) {
		t.Fatalf("got %#v, want %#v", summary.SiteCounts, want)
	}
}

func TestLoadRepositoryResourcesWithBOMHeader(t *testing.T) {
	resources, err := LoadResources(filepath.Join("..", "..", "assets", "resourceId.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) < 30 {
		t.Fatalf("got %d resources, want at least 30", len(resources))
	}
	if resource, ok := resources[5]; !ok || resource.Name == "" || resource.Image == nil {
		t.Fatalf("resource 5 was not decoded: %#v", resource)
	}
}

func TestGenerateMapArtifacts(t *testing.T) {
	root := filepath.Join("..", "..")
	outputDir := t.TempDir()
	drops := []Drop{
		{SiteID: 5, ResourceID: 1, PositionX: 1, PositionZ: 2},
		{SiteID: 5, ResourceID: 5, PositionX: 1, PositionZ: 2},
		{SiteID: 5, ResourceID: 20, PositionX: 3, PositionZ: 4},
	}
	result, err := Generate(
		drops,
		filepath.Join(root, "assets", "resourceId.csv"),
		filepath.Join(root, "assets", "NotoSansSC-Regular.ttf"),
		outputDir,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.MapFiles) != 1 || result.Summary.TotalDrops != len(drops) {
		t.Fatalf("unexpected result: %#v", result)
	}
	file, err := os.Open(result.MapFiles[0])
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	decoded, err := png.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != canvasWidth || decoded.Bounds().Dy() != canvasHeight {
		t.Fatalf("got PNG bounds %v, want %dx%d", decoded.Bounds(), canvasWidth, canvasHeight)
	}
	if _, err := os.Stat(result.RareFile); err != nil {
		t.Fatalf("rare resource file missing: %v", err)
	}
	leftovers, err := filepath.Glob(filepath.Join(outputDir, ".*.tmp"))
	if err != nil {
		t.Fatal(err)
	}
	if len(leftovers) != 0 {
		t.Fatalf("atomic writes left temporary files: %v", leftovers)
	}
}
