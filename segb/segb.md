# The SEGB file format

The SEGB file format is a binary file format that stores Biome data streams on Apple devices. These streams are integral to the devices functionality (mainly usage data) and are used by the device to determine the users habits and preferences. The SEGB file format is a proprietary format and is not publicly documented.

Thanks to the efforts of the [
CCL Solutions Group](https://github.com/cclgroupltd) team, both versions of the format have been reverse engineered.

> **Note:** All numbers are assumed to be in little-endian format unless otherwise specified

## Version 2

### Header

The header stores the necessary info for the SEGB file, mainly things for parsing. It does not, however, contain any important data.

- Length: `0x20` bytes (32 bytes)
- Offset: `0x00`

| Offset | Size (bytes) | Type    | Description                                                                         |
|--------|--------------|---------|-------------------------------------------------------------------------------------|
| 0x00   | 4            | char[4] | File magic (SEGB)                                                                   |
| 0x04   | 4            | int32   | Entry count                                                                         |
| 0x08   | 8            | double  | Cocoa timestamp of creation (seconds since 2001-01-01 00:00:00)                     |
| 0x16   | 12           | unknown | Internal parsing meta [(according to Cellebrite)](https://arc.net/l/quote/vlvqmokm) |
| 0x28   | 4            | n/a     | Padding                                                                             |

### Trailer

The trailer contains a record for each entry in the file.

- Length: `ENTRY_SIZE * ENTRY_COUNT` bytes
- Offset: `FILE_SIZE - (ENTRY_SIZE * ENTRY_COUNT)`

#### Record

Each record is a constant size and contains a reference to the real data.

- Length: `0x10` bytes (16 bytes)
- Offset (from start of trailer): `RECORD_SIZE * RECORD_INDEX`

| Offset | Size (bytes) | Type               | Description                                                     |
|--------|--------------|--------------------|-----------------------------------------------------------------|
| 0x00   | 4            | int32              | Entry offset (relative to end of header, so + HEADER_SIZE)      |
| 0x04   | 4            | int32 (EntryState) | Entry state                                                     |
| 0x08   | 8            | double             | Cocoa timestamp of creation (seconds since 2001-01-01 00:00:00) |

> **enum EntryState** <br>
> `0x01` Written <br>
> `0x03` Deleted <br>
> `0x04` Unknown <br>

### Entry

- Length: `ENTRY_OFFSET - POSITION` bytes
- Offset: `POSITION`

| Offset | Size (bytes)    | Type    | Description            |
|--------|-----------------|---------|------------------------|
| 0x00   | 4               | uint32  | CRC of main entry data |
| 0x04   | 4               | unknown | Unknown data           |
| 0x08   | `LENGTH` - 8    | any     | Entry data             |

> **Note:** Each entry is aligned to 0x04 bytes, so when seeking to the next entry, you should align the position to the next 0x04 byte boundary.

## Version 1

### Header

The header contains essential information for parsing the SEGB file, including the end of data offset and a file magic identifier. It does not contain any critical data beyond parsing metadata.

- **Length:** `0x38` bytes (56 bytes)
- **Offset:** `0x00`

| Offset | Size (bytes) | Type     | Description                                     |
|--------|--------------|----------|-------------------------------------------------|
| 0x00   | 4            | uint32   | End of data offset (where entry data ends)      |
| 0x04   | 48           | Unknown  | Unknown data                                    |
| 0x34   | 4            | char[4]  | File magic (`SEGB`)                             |

### Entries

Entries are stored sequentially after the header and continue until the end of data offset specified in the header. Each entry consists of a fixed-size record header, followed by variable-length data, and is aligned to an 8-byte boundary.

- **Length:** `0x20` bytes (32 bytes)

| Offset | Size (bytes)   | Type   | Description                                         |
|--------|----------------|--------|-----------------------------------------------------|
| 0x00   | 4              | int32  | Entry length (length of the data section in bytes)  |
| 0x04   | 4              | int32  | Entry state                                         |
| 0x08   | 8              | double | Timestamp 1 (seconds since 2001-01-01 00:00:00 UTC) |
| 0x10   | 8              | double | Timestamp 2 (seconds since 2001-01-01 00:00:00 UTC) |
| 0x18   | 4              | uint32 | CRC32 checksum of the data section                  |
| 0x1C   | 4              | int32  | Unknown                                             |
| 0x20   | `ENTRY_LENGTH` | any    | Entry data                                          |

> **enum EntryState**  
> `0x01` Written  
> `0x03` Deleted  
> `0x04` Unknown

> **Note:** Each entry is aligned to 8 bytes. When seeking to the next entry, align the position to the next 8-byte boundary.