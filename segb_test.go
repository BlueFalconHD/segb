package segb

import (
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

var expectedEntryData = []string{
	"Here's to the crazy ones.",
	"The misfits.",
	"The rebels.",
}

func SetupTestFiles() {
	cmd := exec.Command("python3", "generate_test_files.py")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to generate test files: %v", err)
	}
}

func RemoveTestFiles() {
	err := os.Remove("segb_version1.bin")
	if err != nil {
		log.Fatalf("Failed to remove test file: %v", err)
	}

	err = os.Remove("segb_version2.bin")
	if err != nil {
		log.Fatalf("Failed to remove test file: %v", err)
	}
}

func CheckForEntries(t *testing.T, entries []Entry) {
	if len(entries) != len(expectedEntryData) {
		t.Fatalf("len(entries) = %d; want %d", len(entries), len(expectedEntryData))
	}

	for i, entry := range entries {
		if entry.ID != i {
			t.Errorf("entry.ID = %d; want %d", entry.ID, i)
		}
		if string(entry.Data) != expectedEntryData[i] {
			t.Errorf("entry.Data = %s; want %s", entry.Data, expectedEntryData[i])
		}
	}
}

func TestCocoaTimestampToTime(t *testing.T) {
	// Test the CocoaTimestampToTime function
	// Test with a known timestamp
	timestamp := 0.0
	expectedTime := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
	resultTime := CocoaTimestampToTime(timestamp)
	if resultTime != expectedTime {
		t.Errorf("CocoaTimestampToTime(%f) = %v; want %v", timestamp, resultTime, expectedTime)
	}
}

func TestDetectVersion(t *testing.T) {

	SetupTestFiles()
	defer RemoveTestFiles()

	// Test the DetectVersion function
	// Test with a Version 1 SEGB file
	fileV1, err := os.Open("segb_version1.bin")
	if err != nil {
		t.Fatal(err)
	}

	// Create a io.ReadSeeker from the file
	version, err := DetectVersion(fileV1)
	if err != nil {
		t.Fatal(err)
	}

	if version != SEGB_VERSION_1 {
		t.Errorf("DetectVersion() = %v; want %v", version, SEGB_VERSION_1)
	}

	// Test with a Version 2 SEGB file
	fileV2, err := os.Open("segb_version2.bin")
	if err != nil {
		t.Fatal(err)
	}

	// Create a io.ReadSeeker from the file
	version, err = DetectVersion(fileV2)
	if err != nil {
		t.Fatal(err)
	}

	if version != SEGB_VERSION_2 {
		t.Errorf("DetectVersion() = %v; want %v", version, SEGB_VERSION_2)
	}
}

func TestDecode(t *testing.T) {

	SetupTestFiles()
	defer RemoveTestFiles()

	fileV1, err := os.Open("segb_version1.bin")
	if err != nil {
		t.Fatal(err)
	}

	filev2, err := os.Open("segb_version2.bin")
	if err != nil {
		t.Fatal(err)
	}

	decoded, err := Decode(fileV1)
	if err != nil {
		t.Fatal(err)
	}

	if len(decoded.Entries) != 3 {
		t.Errorf("Decode() returned %d entries; want 3", len(decoded.Entries))
	}

	if decoded.Version != SEGB_VERSION_1 {
		t.Errorf("Decode() returned version %v; want %v", decoded.Version, SEGB_VERSION_1)
	}

	// Check the entries
	CheckForEntries(t, decoded.Entries)

	decoded, err = Decode(filev2)
	if err != nil {
		t.Fatal(err)
	}

	if len(decoded.Entries) != 3 {
		t.Errorf("Decode() returned %d entries; want 3", len(decoded.Entries))
	}

	if decoded.Version != SEGB_VERSION_2 {
		t.Errorf("Decode() returned version %v; want %v", decoded.Version, SEGB_VERSION_2)
	}

	// Check the entries
	CheckForEntries(t, decoded.Entries)
}
