package ps3

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/binary"
	"os"
)

var (
	SYSCON_MANAGER_KEY     = []byte{0xd4, 0x13, 0xb8, 0x96, 0x63, 0xe1, 0xfe, 0x9f, 0x75, 0x14, 0x3d, 0x3b, 0xb4, 0x56, 0x52, 0x74}
	KEYGEN_KEY             = []byte{0x6b, 0x1a, 0xce, 0xa2, 0x46, 0xb7, 0x45, 0xfd, 0x8f, 0x93, 0x76, 0x3b, 0x92, 0x05, 0x94, 0xcd, 0x53, 0x48, 0x3b, 0x82}
	SAVEGAME_PARAM_SFO_KEY = []byte{0x0c, 0x08, 0x00, 0x0e, 0x09, 0x05, 0x04, 0x04, 0x0d, 0x01, 0x0f, 0x00, 0x04, 0x06, 0x02, 0x02, 0x09, 0x06, 0x0d, 0x03}
)

func hmacDigest(key []byte, data []byte) []byte {
	h := hmac.New(sha1.New, key)
	return h.Sum(data)
}

func MakePfd(version uint64, sfo []byte, dir string) error {
	// these are normally random, but we can just null them out
	pfHeaderIv := make([]byte, 16)
	pfKeyOrig := make([]byte, 20)

	pfKey := pfKeyOrig
	if version == 4 {
		copy(pfKey, hmacDigest(KEYGEN_KEY, pfKeyOrig))
	}

	pfIndexSize := 1 // normally 57, we only need 1
	pfEntrySize := 1 // normally 114, we only need 1

	sfoFilename := "PARAM.SFO"

	// the only protected file entry, for our PARAM.SFO
	pfEntries := new(bytes.Buffer)
	binary.Write(pfEntries, binary.BigEndian, pfIndexSize)
	pfEntries.WriteString(sfoFilename)

	pfEntries.Write(make([]byte, 7))
	pfEntries.Write(make([]byte, 64))
	pfEntries.Write(hmacDigest(SAVEGAME_PARAM_SFO_KEY, sfo))
	pfEntries.Write(make([]byte, 20)) // console id hash
	pfEntries.Write(make([]byte, 20)) // disc key hash
	pfEntries.Write(make([]byte, 20)) // account id hash
	pfEntries.Write(make([]byte, 40)) // reserved
	binary.Write(pfEntries, binary.BigEndian, uint64(len(sfo)))

	// protected file index
	pfIndex := new(bytes.Buffer)
	binary.Write(pfIndex, binary.BigEndian, pfIndexSize)
	binary.Write(pfIndex, binary.BigEndian, pfEntrySize) // reserved entries
	binary.Write(pfIndex, binary.BigEndian, pfEntrySize) // used entries
	binary.Write(pfIndex, binary.BigEndian, uint64(0))   // used entries

	h := hmac.New(sha1.New, pfKey)
	h.Write([]byte(sfoFilename))
	// signature doesn't include next entry index or the padding after file name
	// only one pf entry, so only one sig in the sig table

	pfEntrySigTable := h.Sum(pfEntries.Bytes()[80:])

	// signature for pf index
	pfIndexSig := hmacDigest(pfKey, pfIndex.Bytes())

	// signature for pf entry sig table
	pfEntrySigTableSig := hmacDigest(pfKey, pfEntrySigTable)

	pfHeader := new(bytes.Buffer)
	pfHeader.Write(pfEntrySigTableSig)
	pfHeader.Write(pfIndexSig)
	pfHeader.Write(pfKeyOrig)
	pfHeader.Write(make([]byte, 4))

	block, err := aes.NewCipher(SYSCON_MANAGER_KEY)
	if err != nil {
		return err
	}
	cbc := cipher.NewCBCEncrypter(block, pfHeaderIv)
	encPfHeader := make([]byte, len(pfHeader.Bytes()))
	cbc.CryptBlocks(encPfHeader, pfHeader.Bytes())

	f, err := os.OpenFile("PARAM.PFD", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("\000\000\000\000PFDB")
	binary.Write(f, binary.BigEndian, version)

	f.Write(pfHeaderIv)
	f.Write(pfHeader.Bytes())

	f.Write(pfIndex.Bytes())
	f.Write(pfEntries.Bytes())
	f.Write(pfEntrySigTable)

	return nil
}
