// Package v1 provides functionality to read and parse SEGB version 1 files.
package v1

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
)

const (
	// FileMagic is the expected magic number at the end of the header.
	FileMagic = "SEGB"
)

// EntryState represents the state of an entry.
type EntryState int32

const (
	EntryStateWritten EntryState = 0x01
	EntryStateDeleted EntryState = 0x03
	EntryStateUnknown EntryState = 0x04
)

// Header represents the header of a SEGB version 1 file.
type Header struct {
	EndOfDataOffset int32    // Offset where entry data ends.
	_               [48]byte // Unknown data (purpose not yet identified).
	Magic           [4]byte  // File magic number, should be "SEGB".
}

// IsValidMagic checks if the magic number matches the expected value.
func (h *Header) IsValidMagic() bool {
	return string(h.Magic[:]) == FileMagic
}

// Entry represents an entry in a SEGB version 1 file.
type Entry struct {
	ID          int32      // Entry ID.
	Length      int32      // Length of the data section in bytes.
	State       EntryState // State of the entry.
	Timestamp1  float64    // First timestamp (Cocoa timestamp).
	Timestamp2  float64    // Second timestamp (Cocoa timestamp).
	CRCChecksum uint32     // CRC32 checksum of the data section.
	Unknown     int32      // Unknown field.
	Data        []byte     // Entry data.

	// Additional fields for convenience.
	Offset int64 // Offset of the entry in the file.
}

// VerifyCRC calculates the CRC32 checksum of the entry data and compares it with the stored checksum.
func (e *Entry) VerifyCRC() bool {
	calculatedCRC := crc32.Checksum(e.Data, crc32.IEEETable)
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

// ReadEntry reads an entry from the provided stream.
func ReadEntry(stream io.ReadSeeker, idx int32) (*Entry, error) {
	entry := &Entry{}
	// Record the current offset
	offset, err := stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	entry.Offset = offset

	// Read the fixed-size entry header
	err = binary.Read(stream, binary.LittleEndian, &entry.Length)
	if err != nil {
		return nil, err
	}
	err = binary.Read(stream, binary.LittleEndian, &entry.State)
	if err != nil {
		return nil, err
	}
	err = binary.Read(stream, binary.LittleEndian, &entry.Timestamp1)
	if err != nil {
		return nil, err
	}
	err = binary.Read(stream, binary.LittleEndian, &entry.Timestamp2)
	if err != nil {
		return nil, err
	}
	err = binary.Read(stream, binary.LittleEndian, &entry.CRCChecksum)
	if err != nil {
		return nil, err
	}
	err = binary.Read(stream, binary.LittleEndian, &entry.Unknown)
	if err != nil {
		return nil, err
	}

	// set ID
	entry.ID = idx

	// Read the variable-length data section
	entry.Data = make([]byte, entry.Length)
	_, err = io.ReadFull(stream, entry.Data)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// ReadSegb reads and parses a SEGB version 1 file from the provided stream.
// It returns the header, a slice of entries, and an error if any.
func ReadSegb(stream io.ReadSeeker) (*Header, []*Entry, error) {
	// Read the header
	header, err := ReadHeader(stream)
	if err != nil {
		return nil, nil, err
	}

	// Verify the magic number
	if !header.IsValidMagic() {

		return nil, nil, fmt.Errorf("invalid magic number: %s", string(header.Magic[:]))
	}

	// Initialize an empty slice to hold entries
	entries := []*Entry{}

	idx := int32(0)

	// Entries start immediately after the header
	for {
		// Get the current position
		currentPosition, err := stream.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, nil, err
		}

		// Check if we've reached the end of data
		if int32(currentPosition) >= header.EndOfDataOffset {
			break
		}

		// Read the next entry
		entry, err := ReadEntry(stream, idx)
		if err != nil {
			return nil, nil, err
		}
		entries = append(entries, entry)

		// Align to 8-byte boundary
		positionAfterEntry, err := stream.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, nil, err
		}
		padding := (8 - (positionAfterEntry % 8)) % 8
		if padding > 0 {
			_, err = stream.Seek(padding, io.SeekCurrent)
			if err != nil {
				return nil, nil, err
			}
		}

		idx++
	}

	return header, entries, nil
}
