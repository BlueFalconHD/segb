// Package v2 provides functionality to read and parse SEGB v2 files.
package v2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"sort"
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

const (
	// FileMagic is the expected magic number at the beginning of the file.
	FileMagic = "SEGB"
	// TrailerRecordSize is the size in bytes of each trailer record.
	TrailerRecordSize = 16
)

// EntryState represents the state of an entry.
type EntryState int32

const (
	EntryStateWritten EntryState = 0x01
	EntryStateDeleted EntryState = 0x03
	EntryStateUnknown EntryState = 0x04
)

// Header represents the header of a SEGB file.
type Header struct {
	Magic             [4]byte  // File magic number, should be "SEGB"
	EntryCount        int32    // Number of entries in the file
	CreationTimestamp float64  // Creation timestamp (Cocoa timestamp)
	UnknownPadding    [16]byte // Padding or reserved bytes, purpose unknown
}

// MagicString returns the magic number as a string.
func (h *Header) MagicString() string {
	return string(h.Magic[:])
}

// IsValidMagic checks if the magic number matches the expected value.
func (h *Header) IsValidMagic() bool {
	return h.MagicString() == FileMagic
}

// Record represents a trailer record in a SEGB file.
type Record struct {
	Offset            int32      // Offset of the entry data from the start of entries
	State             EntryState // State of the entry
	CreationTimestamp float64    // Creation timestamp (Cocoa timestamp)
}

// Entry represents an entry in a SEGB file.
type Entry struct {
	ID                uint32     // Entry identifier
	State             EntryState // State of the entry
	CreationTimestamp float64    // Creation timestamp (Cocoa timestamp)

	CRCChecksum uint32  // CRC32 checksum of the entry data
	Unknown     [4]byte // Unknown 4 bytes
	Data        []byte  // Entry data (NB: due to some kinks with alignment, this might contain extra zero bytes. Trim as needed)

	RawData []byte // Raw data including CRCChecksum and Unknown fields
}

// VerifyCRC calculates the CRC32 checksum of the entry data and compares it with the stored checksum.
func (e *Entry) VerifyCRC() bool {
	// Exclude the CRCChecksum and Unknown fields (first 8 bytes)
	dataToCheck := e.RawData[8:]
	calculatedCRC := crc32.Checksum(dataToCheck, crc32.IEEETable)
	return e.CRCChecksum == calculatedCRC
}

// ReadHeader reads the header from the provided stream.
func ReadHeader(stream io.ReadSeeker) (*Header, error) {
	header := &Header{}
	err := binary.Read(stream, binary.LittleEndian, header)
	if err != nil {
		return nil, err
	}
	return header, nil
}

// ReadRecord reads a trailer record from the provided stream.
func ReadRecord(stream io.ReadSeeker) (*Record, error) {
	record := &Record{}
	err := binary.Read(stream, binary.LittleEndian, record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// ReadSegb reads and parses a SEGB version 2 file from the provided stream.
// It returns the header, a slice of records, a slice of entries, and an error if any.
func ReadSegb(stream io.ReadSeeker) (*Header, []*Record, []*Entry, error) {
	// Read the header
	header, err := ReadHeader(stream)
	if err != nil {
		return nil, nil, nil, err
	}

	// Verify the magic number
	if !header.IsValidMagic() {
		return nil, nil, nil, fmt.Errorf("invalid magic number: %s", header.MagicString())
	}

	// Seek to the start of the trailer (list of records)
	trailerSize := TrailerRecordSize * int64(header.EntryCount)
	trailerOffset, err := stream.Seek(-trailerSize, io.SeekEnd)
	if err != nil {
		return nil, nil, nil, err
	}

	// Read the trailer records
	records := make([]*Record, header.EntryCount)
	for i := 0; i < int(header.EntryCount); i++ {
		record, err := ReadRecord(stream)
		if err != nil {
			return nil, nil, nil, err
		}
		records[i] = record
	}

	// Sort records by Offset
	sort.Slice(records, func(i, j int) bool {
		return records[i].Offset < records[j].Offset
	})

	// Read entries
	entries := make([]*Entry, 0, len(records))
	for idx, record := range records {
		if record.State == EntryStateUnknown {
			continue
		}

		// Calculate the start position of the entry
		entryStart := int64(binary.Size(Header{})) + int64(record.Offset)

		// Calculate the length of the entry data
		var entryLength int64
		if idx < len(records)-1 {
			// Not the last record, so entry length is up to the next entry
			nextRecord := records[idx+1]
			entryLength = int64(nextRecord.Offset) - int64(record.Offset)
		} else {
			// Last record, entry length is up to the start of the trailer
			entryLength = trailerOffset - entryStart
		}

		if entryLength <= 0 {
			return nil, nil, nil, fmt.Errorf("invalid entry length")
		}

		// Seek to the entry start position
		_, err = stream.Seek(entryStart, io.SeekStart)
		if err != nil {
			return nil, nil, nil, err
		}

		// Read the entry data
		entryData := make([]byte, entryLength)
		_, err = io.ReadFull(stream, entryData)
		if err != nil {
			return nil, nil, nil, err
		}

		// Parse the entry
		entry := &Entry{}
		if len(entryData) < 8 {
			return nil, nil, nil, fmt.Errorf("entry data too short")
		}
		buf := bytes.NewReader(entryData[:8])
		// Read CRCChecksum and Unknown fields
		err = binary.Read(buf, binary.LittleEndian, &entry.CRCChecksum)
		if err != nil {
			return nil, nil, nil, err
		}
		err = binary.Read(buf, binary.LittleEndian, &entry.Unknown)
		if err != nil {
			return nil, nil, nil, err
		}

		entry.ID = uint32(idx)
		entry.State = record.State
		entry.CreationTimestamp = record.CreationTimestamp
		entry.Data = bytes.TrimRight(entryData[8:], "\x00") // Data after CRCChecksum and Unknown fields, trim padding
		entry.RawData = entryData

		entries = append(entries, entry)

		// Handle alignment padding by seeking to the next 4-byte boundary
		currentPosition := entryStart + entryLength
		alignment := (4 - (currentPosition % 4)) % 4
		if alignment > 0 {
			_, err = stream.Seek(alignment, io.SeekCurrent)
			if err != nil {
				return nil, nil, nil, err
			}
		}
	}

	return header, records, entries, nil
}
