package archive

import (
	"archive/zip"
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"github.com/Zaprit/LBPSearch/pkg/archive/xxtea"
	"io"
	"path"
	"slices"
	"strconv"
	"strings"
	"sync"
)

var HashinateKey = []byte{
	0x2A, 0xFD, 0xA3, 0xCA, 0x86, 0x02, 0x19, 0xB3,
	0xE6, 0x8A, 0xFF, 0xCC, 0x82, 0xC7, 0x6B, 0x8A,
	0xFE, 0x0A, 0xD8, 0x13, 0x5F, 0x60, 0x47, 0x5B,
	0xDF, 0x5D, 0x37, 0xBC, 0x57, 0x1C, 0xB5, 0xE7,
	0x96, 0x75, 0xD5, 0x28, 0xA2, 0xFA, 0x90, 0xED,
	0xDF, 0xA3, 0x45, 0xB4, 0x1F, 0xF9, 0x1F, 0x25,
	0xE7, 0x42, 0x45, 0x3B, 0x2B, 0xB5, 0x3E, 0x16,
	0xC9, 0x58, 0x19, 0x7B, 0xE7, 0x18, 0xC0, 0x80,
}

var TeaKey = []byte{
	0x01, 0xB7, 0x0C, 0xBD, 0x14, 0x96, 0x07, 0xD6, 0x07, 0xF9, 0x4D, 0xD5, 0x10, 0xDB, 0x8C, 0xA0,
}

const ChunkSize = 0x240000

type ArchiveEntry struct {
	SHA1   string
	Offset uint32
	Size   uint32
}

func MakeSaveArchive(rev ResourceRevision, rootLevel []byte, hashes *sync.Map, backupZip zip.Writer, backupDir string) error {

	arc := new(bytes.Buffer)
	var entries []ArchiveEntry

	for hash, resource := range hashes.Range {
		resource, ok := resource.(*ResourceReference)
		if !ok {
			continue
		}

		offset := arc.Len()

		_, err := io.Copy(arc, resource.file)
		if err != nil {
			return err
		}

		info, err := resource.file.Stat()
		if err != nil {
			return err
		}

		entries = append(entries, ArchiveEntry{
			SHA1:   hash.(string),
			Offset: uint32(offset),
			Size:   uint32(info.Size()),
		})
	}

	// align to 4 byte boundary
	pad := arc.Len() % 4
	if pad != 0 {
		pad = 4 - pad
	}

	arc.WriteString(strings.Repeat("\x00", pad))

	// save key
	binary.Write(arc, binary.BigEndian, rev.Head)
	binary.Write(arc, binary.BigEndian, rev.BranchID)
	binary.Write(arc, binary.BigEndian, rev.BranchRevision)
	binary.Write(arc, binary.BigEndian, uint32(1))  // localUserID
	arc.WriteString(strings.Repeat("\x00", 40))     // deprecated1 int32[10]
	binary.Write(arc, binary.BigEndian, uint32(0))  // copied
	binary.Write(arc, binary.BigEndian, uint32(29)) // root type value, SLOT_LIST
	arc.WriteString(strings.Repeat("\x00", 12))     // deprecated2 int32[3]
	arc.Write(rootLevel)
	arc.WriteString(strings.Repeat("\x00", 40)) // deprecated3 int32[10]

	// fat entries
	for _, entry := range entries {
		arc.WriteString(entry.SHA1)
		binary.Write(arc, binary.BigEndian, entry.Offset)
		binary.Write(arc, binary.BigEndian, entry.Size)
	}

	// hashinate, to be written later
	hashinate_offset := arc.Len()
	arc.WriteString(strings.Repeat("\x00", 20))

	binary.Write(arc, binary.BigEndian, uint32(len(entries)))
	arc.WriteString("FAR4")

	mac := hmac.New(sha1.New, HashinateKey)
	mac.Write(arc.Bytes())

	for i, b := range mac.Sum(nil) {
		arc.Bytes()[hashinate_offset+i] = b
	}

	i := 0
	for chunk := range slices.Chunk(arc.Bytes(), ChunkSize) {
		xxtea.Encrypt(chunk, TeaKey)

		entry, err := backupZip.Create(path.Join(backupDir, strconv.Itoa(i)))
		if err != nil {
			return err
		}
		_, err = io.Copy(entry, arc)
		if err != nil {
			return err
		}
	}
	return nil
}
