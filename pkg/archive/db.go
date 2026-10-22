package archive

type LevelType uint

const (
	Cooperative LevelType = iota
	Verses
	Cutscene
)

type GameVersion uint

const (
	LBP1 GameVersion = iota
	LBP2
	LBP3
)

func (gv GameVersion) GetTitle() string {
	switch gv {
	case LBP1:
		return "LittleBigPlanet™"
	case LBP2:
		return "LittleBigPlanet™2"
	case LBP3:
		return "LittleBigPlanet™3"
	default:
		panic("Unknown GameVersion")
	}
}

func (gv GameVersion) GetShortTitle() string {
	switch gv {
	case LBP1:
		return "LBP1"
	case LBP2:
		return "LBP2"
	case LBP3:
		return "LBP3"
	default:
		panic("Unknown GameVersion")
	}
}

func (gv GameVersion) GetTitleID() string {
	switch gv {
	case LBP1:
		return "BCES00141"
	case LBP2:
		return "BCES00850"
	case LBP3:
		return "BCES01663"
	default:
		panic("Unknown GameVersion")
	}
}

func (gv GameVersion) GetLatestRevision() ResourceRevision {
	switch gv {
	case LBP1:
		return ResourceRevision{
			Head:           0x272,
			BranchID:       0x4c44,
			BranchRevision: 0x17,
		}
	case LBP2:
		return ResourceRevision{
			Head:           0x3f8,
			BranchID:       0x0,
			BranchRevision: 0x0,
		}
	case LBP3:
		return ResourceRevision{
			Head:           0x21803f9,
			BranchID:       0x0,
			BranchRevision: 0x0,
		}
	default:
		panic("Unknown GameVersion")
	}
}

type ResourceType uint

const (
	SHA1 ResourceType = iota
	GUID
)

type ResourceDependency struct {
	ResType uint32
	Desc    ResourceDescriptor
}

type ResourceDescriptor struct {
	Type ResourceType
	Guid uint32
	Sha  []byte
}

type SlotInfo struct {
	Name              string
	Description       string
	NPHandle          string
	RootLevel         [20]byte
	Icon              ResourceDescriptor
	Game              GameVersion
	InitiallyLocked   bool
	IsSubLevel        bool
	BackgroundGUID    *uint32
	Shareable         bool
	AuthorLabels      []uint32
	LevelType         LevelType
	MinPlayers        *uint8
	MaxPlayers        *uint8
	IsAdventurePlanet bool
}
