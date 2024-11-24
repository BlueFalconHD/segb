package segb

import (
	"errors"
	"fmt"
	v1 "github.com/bluefalconhd/segb/v1"
	v2 "github.com/bluefalconhd/segb/v2"
	"hash/crc32"
	"io"
	"time"
)

func PrettyHexdump(data []byte) {
	for i := 0; i < len(data); i += 16 {
		fmt.Printf("%08x: ", i)
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				fmt.Printf("%02x ", data[i+j])
			} else {
				fmt.Print("   ")
			}
		}
		fmt.Print(" ")
		for j := 0; j < 16; j++ {
			if i+j < len(data) {
				if data[i+j] >= 32 && data[i+j] <= 126 {
					fmt.Printf("%c", data[i+j])
				} else {
					fmt.Print(".")
				}
			}
		}
		fmt.Println()
	}
}

type SegbVersion int

const (
	NONE SegbVersion = iota
	SEGB_VERSION_1
	SEGB_VERSION_2
)

var ErrUnsupportedVersion = errors.New("unsupported version")

func Decode(stream io.ReadSeeker) (Segb, error) {

	// Detect the version of the SEGB file
	v, err := DetectVersion(stream)
	if err != nil {
		return Segb{}, err
	}

	// Re-seek to the beginning of the file (this took me so long to realize)
	_, err = stream.Seek(0, io.SeekStart)
	if err != nil {
		return Segb{}, err
	}

	switch v {
	case SEGB_VERSION_1:
		header, entries, err := v1.ReadSegb(stream)
		if err != nil {
			return Segb{}, err
		}
		return V1ToStandardSegb(header, entries), nil
	case SEGB_VERSION_2:
		header, _, entries, err := v2.ReadSegb(stream)
		if err != nil {
			return Segb{}, err
		}

		return V2ToStandardSegb(header, entries), nil
	default:
		// Return an error if the version is not supported
		return Segb{}, ErrUnsupportedVersion
	}
}

func DetectVersion(stream io.ReadSeeker) (SegbVersion, error) {
	// Buffer to hold the magic string
	magic := make([]byte, 4)

	// Check for SEGBv2: 'SEGB' @ 0x00
	_, err := stream.Seek(0x00, io.SeekStart)
	if err != nil {
		return NONE, err
	}
	_, err = stream.Read(magic)
	if err != nil {
		return NONE, err
	}
	if string(magic) == "SEGB" {
		return SEGB_VERSION_2, nil
	}

	// Check for SEGBv1: 'SEGB' @ 0x34
	_, err = stream.Seek(0x34, io.SeekStart)
	if err != nil {
		return NONE, err
	}
	_, err = stream.Read(magic)
	if err != nil {
		return NONE, err
	}
	if string(magic) == "SEGB" {
		return SEGB_VERSION_1, nil
	}

	// If neither version is detected, return NONE
	return NONE, nil
}

func CocoaTimestampToTime(timestamp float64) time.Time {
	return time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(timestamp) * time.Second)
}

func V2EntryStateToStandardState(e v2.EntryState) EntryState {
	switch e {
	case v2.EntryStateWritten:
		return EntryStateWritten
	case v2.EntryStateDeleted:
		return EntryStateDeleted
	default:
		return EntryStateUnknown
	}
}

func V1EntryStateToStandardState(e v1.EntryState) EntryState {
	switch e {
	case v1.EntryStateWritten:
		return EntryStateWritten
	case v1.EntryStateDeleted:
		return EntryStateDeleted
	default:
		return EntryStateUnknown
	}
}

func V1ToStandardSegb(header *v1.Header, entries []*v1.Entry) Segb {
	oldestTime := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)

	standardEntries := make([]Entry, len(entries))
	for i, entry := range entries {

		// Calculate the creation time
		creationTime := CocoaTimestampToTime(entry.Timestamp1)
		if creationTime.Before(oldestTime) {
			oldestTime = creationTime
		}

		standardEntries[i] = Entry{
			ID:       int(entry.ID),
			State:    V1EntryStateToStandardState(entry.State),
			Created:  CocoaTimestampToTime(entry.Timestamp1),
			Data:     entry.Data,
			Checksum: entry.CRCChecksum,
		}
	}
	return Segb{
		Version: SEGB_VERSION_1,
		// Creation time is unknown for SEGBv1, so we use the oldest entry creation time
		Created: oldestTime,
		Entries: standardEntries,
	}
}
func V2ToStandardSegb(header *v2.Header, entries []*v2.Entry) Segb {

	standardEntries := make([]Entry, len(entries))
	for i, entry := range entries {
		standardEntries[i] = Entry{
			ID:       int(entry.ID),
			State:    V2EntryStateToStandardState(entry.State),
			Created:  CocoaTimestampToTime(entry.CreationTimestamp),
			Data:     entry.Data,
			Checksum: entry.CRCChecksum,
		}
	}

	return Segb{
		Version: SEGB_VERSION_2,
		Created: CocoaTimestampToTime(header.CreationTimestamp),
		Entries: standardEntries,
	}
}

type EntryState int

const (
	EntryStateWritten EntryState = 0x01
	EntryStateDeleted EntryState = 0x03
	EntryStateUnknown EntryState = 0x04
)

// Entry
type Entry struct {
	ID       int
	State    EntryState
	Created  time.Time
	Data     []byte
	Checksum uint32
}

func (e *Entry) CheckCRC() bool {
	return e.Checksum == crc32.Checksum(e.Data, crc32.IEEETable)
}

type Segb struct {
	Version SegbVersion
	Created time.Time
	Entries []Entry
}
