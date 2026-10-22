package archive

func GameVersionFromDB(game int16) GameVersion {
	switch game {
	case 0:
		return LBP1
	case 1:
		return LBP2
	case 2:
		return LBP3
	default:
		panic("invalid game version")
	}
}
