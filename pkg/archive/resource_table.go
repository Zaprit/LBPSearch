package archive

import (
	"encoding/binary"
	"fmt"
	"io"
)

func ParseTable(r io.ReadSeeker) ([]ResourceDependency, error) {

	var tableOffset uint32
	err := binary.Read(r, binary.BigEndian, &tableOffset)
	if err != nil {
		return nil, err
	}
	origOffset, err := r.Seek(0, io.SeekCurrent)

	_, err = r.Seek(int64(tableOffset), io.SeekStart)
	if err != nil {
		return nil, err
	}

	var resCount uint32
	err = binary.Read(r, binary.BigEndian, &resCount)

	dependencies := make([]ResourceDependency, 0, resCount)
	for range resCount {
		var depType byte
		err = binary.Read(r, binary.BigEndian, &depType)
		if err != nil {
			return nil, err
		}

		var rsrcDesc ResourceDescriptor
		switch depType {
		case 0: // Something about a dynamic thermometer level, idk
			_, err := r.Seek(4, io.SeekCurrent)
			if err != nil {
				return nil, err
			}
		case 1:
			sha1 := make([]byte, 20)
			_, err := io.ReadFull(r, sha1)
			if err != nil {
				return nil, err
			}
			rsrcDesc = ResourceDescriptor{Type: SHA1, Sha: sha1}
		case 2:
			var guid uint32
			err = binary.Read(r, binary.BigEndian, &guid)
			if err != nil {
				return nil, err
			}
			rsrcDesc = ResourceDescriptor{Type: GUID, Guid: guid}
		default:
			return nil, fmt.Errorf("unknown dependency type %d", depType)
		}

		var rsrcType uint32
		err = binary.Read(r, binary.BigEndian, &rsrcType)
		if err != nil {
			return nil, err
		}

		dependencies = append(dependencies, ResourceDependency{
			Desc:    rsrcDesc,
			ResType: rsrcType,
		})
	}

	_, err = r.Seek(origOffset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	return dependencies, nil
}
