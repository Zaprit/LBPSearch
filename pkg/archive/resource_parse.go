package archive

import (
	"encoding/binary"
	"fmt"
	"io"
	"slices"
)

type ResourceID struct {
	Method ResourceMethod
	Type   [3]byte
}

func GetResourceHeader(r io.ReadSeeker) (*ResourceID, error) {
	header := new(ResourceID)

	err := binary.Read(r, binary.BigEndian, &header.Type)
	if err != nil {
		return nil, err
	}

	var encoding uint32
	err = binary.Read(r, binary.BigEndian, &encoding)
	if err != nil {
		return nil, err
	}

	var method ResourceMethod
	switch encoding {
	case 'b', 'e':
		rev := ResourceRevision{}
		err := binary.Read(r, binary.BigEndian, &rev.Head)
		if err != nil {
			return nil, err
		}

		var dependencies []ResourceDependency
		if rev.Head >= 0x109 {
			dependencies, err = ParseTable(r)
			if err != nil {
				return nil, err
			}
		}

		if !slices.Equal(header.Type[:], []byte("SMH")) && rev.Head >= 0x271 {
			err := binary.Read(r, binary.BigEndian, &rev.BranchID)
			if err != nil {
				return nil, err
			}

			err = binary.Read(r, binary.BigEndian, &rev.BranchRevision)
			if err != nil {
				return nil, err
			}
		}

		method = ResourceMethod{
			IsEncrypted:  encoding == 'e',
			Revision:     rev,
			Dependencies: dependencies,
		}

	case ' ':

	default:
		return nil, fmt.Errorf("invalid resource header")
	}

	header.Method = method

	return header, nil
}

type ResourceRevision struct {
	Head           uint32
	BranchID       uint16
	BranchRevision uint16
}

func (r ResourceRevision) GetVersion() uint16 {
	return uint16(r.Head & 0xFFFF)
}

func (r ResourceRevision) GetSubVersion() uint16 {
	return uint16((r.Head >> 16) & 0xFFFF)
}

func (r ResourceRevision) IsLBP1() bool {
	return r.Head <= 0x272
}

func (r ResourceRevision) IsLBP3() bool {
	return r.Head>>0x10 != 0
}

func (r ResourceRevision) GetGameVersion() GameVersion {
	if r.IsLBP1() {
		return LBP1
	} else if r.IsLBP3() {
		return LBP3
	} else {
		return LBP2
	}
}

type BinaryResource struct {
}

type ResourceMethodType uint

const (
	MethodNone ResourceMethodType = iota
	MethodBinary
	MethodTexture
)

type ResourceMethod struct {
	Type ResourceMethodType

	IsEncrypted  bool
	Revision     ResourceRevision
	Dependencies []ResourceDependency
}
