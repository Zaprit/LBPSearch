package archive

import (
	"context"
	"encoding/hex"
	"io"
	"log/slog"
	"os"
	"path"
	"sync"

	"github.com/Zaprit/LBPSearch/pkg/storage"
)

// FIXME: a fair warning to all ye who enter, this is a half finished terrible port of lbp_archive_dl that I was working on some time late 2025
// it is not up to my usual standards, half the shit is broken, and nothing should use this ever. I will reimplement the file handling parts in this repo at some point
// https://git.valtek.uk/HugeSpaceship/gocraftworld
// until then uhh, don't use this. There's a thing in ../website_old/download_archive.go that does this by calling the archive_dl binary that can be used until a suitable go version can be produced
// you've been warned

type Downloader struct {
	store storage.ResourceStorageBackend
	//conn  *db.Queries
}

func (dl *Downloader) DownloadBackup(ctx context.Context, levelID int32) error {

	//slot, err := dl.conn.GetSlot(ctx, levelID)
	//if err != nil {
	//	return errors.New("level not found")
	//}
	//
	//slog.Debug("Found Level!", "name", slot.Name, "creator", slot.NpHandle, "game", slot.Game)
	//
	//tempDir, err := os.MkdirTemp("", "lbpsearch-"+strconv.FormatInt(int64(levelID), 16))
	//if err != nil {
	//	return err
	//}
	//
	//resDir := path.Join(tempDir, "res")
	//err = os.Mkdir(path.Join(tempDir, "res"), 0755)
	//if err != nil {
	//	return err
	//}
	//
	//rootLevel := hex.EncodeToString(slot.RootLevel)
	//
	//dsCtx, cf := context.WithCancel(ctx)
	//
	//ds := DownloadSession{
	//	store:      dl.store,
	//	hashes:     new(sync.Map),
	//	tempDir:    tempDir,
	//	resDir:     resDir,
	//	ctx:        dsCtx,
	//	cancelFunc: cf,
	//}
	//
	//ref, err := ds.FetchResourceWithDependencies(rootLevel)
	//if err != nil {
	//	return err
	//}
	//
	//if ref.header.Method.Type != MethodBinary {
	//	return errors.New("rootLevel is not binary, it's possibly corrupted")
	//}
	//revision := ref.header.Method.Revision
	//
	//if GameVersionFromDB(slot.Game) != revision.GetGameVersion() {
	//	slog.Warn("Game version mismatch, writing version in slot, level was most likely moved in later game.",
	//		"slot", levelID,
	//		"slotName", slot.Name,
	//		"dbVersion", GameVersionFromDB(slot.Game),
	//		"dataVersion", revision.GetGameVersion(),
	//	)
	//}
	//
	//levelIDBytes := make([]byte, 4)
	//binary.BigEndian.PutUint32(levelIDBytes, uint32(levelID))
	//backupStr := revision.GetGameVersion().GetTitleID() + GetLevelType(slot) + strings.ToUpper(hex.EncodeToString(levelIDBytes))

	return nil
}

//func GetLevelType(slot db.Slot) string {
//	return "LEVEL" // TODO: check between levels and lbp3 adventure backups
//}

type DownloadSession struct {
	store      storage.ResourceStorageBackend
	hashes     *sync.Map
	tempDir    string
	resDir     string
	ctx        context.Context
	cancelFunc context.CancelFunc
}

type ResourceReference struct {
	header *ResourceID
	file   *os.File
}

func (ds *DownloadSession) FetchResourceWithDependencies(hash string) (*ResourceReference, error) {
	if ref, exists := ds.hashes.Load(hash); exists { // Don't re-download
		return ref.(*ResourceReference), nil
	}

	resReader, err := ds.store.GetResource(ds.ctx, storage.FormatURL(hash))
	if err != nil {
		return nil, err
	}
	defer resReader.Close()

	f, err := os.OpenFile(path.Join(ds.resDir, hash), os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	res, err := GetResourceHeader(f)
	if err != nil {
		return nil, err
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return nil, err
	}

	ref := ResourceReference{
		header: res,
		file:   f,
	}
	ds.hashes.Store(hash, &ref)

	wg := new(sync.WaitGroup)

	if res.Method.Type == MethodBinary {
		for _, dependency := range res.Method.Dependencies {
			if dependency.Desc.Type == SHA1 {
				go func() {
					wg.Add(1)
					_, err = ds.FetchResourceWithDependencies(hex.EncodeToString(dependency.Desc.Sha))
					if err != nil {
						slog.Error("Failed to fetch resources", "err", err)
						ds.cancelFunc()
					}
					wg.Done()
				}()
			}
		}
	}
	wg.Wait()
	return &ref, nil
}
