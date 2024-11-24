import struct
import binascii
import datetime

# Base date for timestamps (2001-01-01 00:00:00 UTC)
BASE_DATE = datetime.datetime(2001, 1, 1, tzinfo=datetime.timezone.utc)

def get_timestamp(dt):
    """Calculate the timestamp as seconds since the base date."""
    delta = dt - BASE_DATE
    return delta.total_seconds()

# Entries data: Lines from the poem with associated dates
entries_data = [
    {
        'text': "Here's to the crazy ones.",
        'date': datetime.datetime(2007, 1, 9, tzinfo=datetime.timezone.utc)  # iPhone announcement date
    },
    {
        'text': "The misfits.",
        'date': datetime.datetime(2007, 6, 29, tzinfo=datetime.timezone.utc)  # iPhone release date
    },
    {
        'text': "The rebels.",
        'date': datetime.datetime(2011, 10, 5, tzinfo=datetime.timezone.utc)  # Steve Jobs' passing
    }
]

### Version 1 SEGB File Generation ###

# Build the header (56 bytes)
header_v1 = bytearray(0x38)
# File magic at offset 0x34
header_v1[0x34:0x38] = b'SEGB'

entries_v1 = []
current_offset_v1 = 0x38  # Start after the header

for entry_data in entries_data:
    text = entry_data['text']
    dt = entry_data['date']
    timestamp = get_timestamp(dt)

    # Encode the text and calculate lengths
    data_bytes = text.encode('utf-8')
    data_length = len(data_bytes)

    # Build the entry header (32 bytes)
    entry_header = bytearray(0x20)
    # Entry length
    entry_header[0x00:0x04] = struct.pack('<I', data_length)
    # Entry state (Written)
    entry_header[0x04:0x08] = struct.pack('<i', 0x01)
    # Timestamp 1
    entry_header[0x08:0x10] = struct.pack('<d', timestamp)
    # Timestamp 2 (same as Timestamp 1)
    entry_header[0x10:0x18] = struct.pack('<d', timestamp)
    # CRC32 checksum of the data section
    crc32 = binascii.crc32(data_bytes) & 0xffffffff
    entry_header[0x18:0x1C] = struct.pack('<I', crc32)
    # Unknown field (zeros)
    entry_header[0x1C:0x20] = b'\x00' * 4

    # Combine header and data
    entry = entry_header + data_bytes
    # Align to 8 bytes
    padding_needed = (8 - (len(entry) % 8)) % 8
    entry += b'\x00' * padding_needed

    entries_v1.append(entry)
    current_offset_v1 += len(entry)

# Set the End of data offset in the header
header_v1[0x00:0x04] = struct.pack('<I', current_offset_v1)

# Write the Version 1 SEGB file
with open('segb_version1.bin', 'wb') as f:
    f.write(header_v1 + b''.join(entries_v1))

### Version 2 SEGB File Generation ###

# Build the header (32 bytes)
header_v2 = bytearray(0x20)
# File magic
header_v2[0x00:0x04] = b'SEGB'
# Entry count
entry_count = len(entries_data)
header_v2[0x04:0x08] = struct.pack('<I', entry_count)
# Timestamp of creation
creation_timestamp = get_timestamp(datetime.datetime.now(tz=datetime.timezone.utc))
header_v2[0x08:0x10] = struct.pack('<d', creation_timestamp)
# Unknown fields (zeros)
header_v2[0x10:0x20] = b'\x00' * 16

entries_v2 = []
entry_offsets_v2 = []
current_offset_v2 = 0  # Offset relative to the end of the header

for entry_data in entries_data:
    text = entry_data['text']
    dt = entry_data['date']
    timestamp = get_timestamp(dt)

    data_bytes = text.encode('utf-8')
    # CRC32 of the data
    crc32 = binascii.crc32(data_bytes) & 0xffffffff
    # Build the entry
    entry = struct.pack('<I', crc32) + b'\x00' * 4 + data_bytes
    # Align to 4 bytes
    padding_needed = (4 - (len(entry) % 4)) % 4
    entry += b'\x00' * padding_needed

    entries_v2.append(entry)
    # Record entry offset and metadata for the trailer
    entry_offsets_v2.append({
        'offset': current_offset_v2,
        'state': 0x01,  # Written
        'timestamp': timestamp
    })
    current_offset_v2 += len(entry)

# Build the trailer
trailer_v2 = bytearray()
for entry_info in entry_offsets_v2:
    trailer_v2 += struct.pack('<I', entry_info['offset'])
    trailer_v2 += struct.pack('<i', entry_info['state'])
    trailer_v2 += struct.pack('<d', entry_info['timestamp'])

# Write the Version 2 SEGB file
with open('segb_version2.bin', 'wb') as f:
    f.write(header_v2 + b''.join(entries_v2) + trailer_v2)