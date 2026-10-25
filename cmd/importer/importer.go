package main

import (
	"context"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/Zaprit/LBPSearch/pkg/model"
)

var DryPath = flag.String("dry", "./dry.db", "path to dry database")

var DBURL = flag.String("url", "host=localhost user=lbpsearch password=lbpsearch dbname=lbpsearch port=5432 sslmode=disable", "destination database dsn")

var NoSlots = flag.Bool("no-slots", false, "skip slot import")
var NoUsers = flag.Bool("no-users", false, "skip user import")

var ExclusionFile = flag.String("exclusion-file", "", "path to exclusion file")

func main() {
	flag.Parse()
	src, err := gorm.Open(sqlite.Open(*DryPath))
	if err != nil {
		panic(err)
	}

	dest, err := gorm.Open(postgres.Open(*DBURL))
	if err != nil {
		panic(err)
	}

	var exclusions ExclusionList
	if *ExclusionFile != "" {
		exclusionBytes, err := os.ReadFile(*ExclusionFile)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(exclusionBytes, &exclusions)
		if err != nil {
			panic(err)
		}
	}

	err = dest.AutoMigrate(&model.Slot{}, &model.User{})
	if err != nil {
		panic(err)
	}
	if !*NoSlots {
		slog.Info("Truncating slots")
		dest.Exec("TRUNCATE TABLE slots;")
	}
	if !*NoUsers {
		slog.Info("Truncating users")
		dest.Exec("TRUNCATE TABLE users;")
	}
	// Give the DB time to actually truncate things
	time.Sleep(5 * time.Second)

	indexStartTime := time.Now()
	slog.Info("Creating Indexes")
	dest.Exec("ALTER TABLE slots ADD COLUMN IF NOT EXISTS textsearchable_index_col tsvector GENERATED ALWAYS AS (to_tsvector('english', coalesce(slots.name, '') || ' ' || coalesce(slots.np_handle, '') || ' ' || coalesce(slots.description, ''))) STORED;")
	dest.Exec("CREATE INDEX IF NOT EXISTS slots_textsearch_idx ON slots USING GIN (textsearchable_index_col);")
	slog.Info("Finished Creating indexes", "indexCreationTime", time.Since(indexStartTime))

	startTime := time.Now()

	totalSlots := 0
	slotImportTime := time.Duration(0)
	if !*NoSlots {
		slotImportTime, totalSlots = MigrateSlots(src, dest)
	}

	totalUsers := 0
	userImportTime := time.Duration(0)
	if !*NoUsers {
		userImportTime, totalUsers = MigrateUsers(src, dest)
	}

	slog.Info("Finished importing", "duration", time.Since(startTime))
	if !*NoSlots {
		slog.Info("Slot Statistics", "totalSlots", totalSlots, "duration", slotImportTime)
	}
	if !*NoUsers {
		slog.Info("User Statistics", "totalUsers", totalUsers, "duration", userImportTime)
	}
}

func MigrateSlots(src *gorm.DB, dest *gorm.DB) (time.Duration, int) {
	startTime := time.Now()
	slots := make([]model.Slot, 0, 1000)
	totalRows := 0
	err := gorm.G[model.ImportSlot](src).Table("slot").FindInBatches(context.Background(), 1000, func(importSlots []model.ImportSlot, batch int) error {
		totalRows += len(importSlots)
		slog.Info("Processing batch of slots", "batch", batch, "rows", len(importSlots), "totalRows", totalRows)
		slots = make([]model.Slot, len(importSlots))

		for i, slot := range importSlots {
			slots[i] = model.Slot{
				ID:                  slot.ID,
				NpHandle:            slot.NpHandle,
				LocationX:           slot.LocationX,
				LocationY:           slot.LocationY,
				Game:                slot.Game,
				Name:                slot.Name,
				Description:         slot.Description,
				RootLevel:           slot.RootLevel,
				Icon:                slot.Icon,
				InitiallyLocked:     slot.InitiallyLocked,
				IsSubLevel:          slot.IsSubLevel,
				IsLBP1Only:          slot.IsLBP1Only,
				Background:          slot.Background,
				Thumbnail:           slot.Thumbnail,
				Shareable:           slot.Shareable,
				AuthorLabels:        slot.AuthorLabels,
				Leveltype:           slot.Leveltype,
				MinPlayers:          slot.MinPlayers,
				MaxPlayers:          slot.MaxPlayers,
				IsAdventurePlanet:   slot.IsAdventurePlanet,
				PS4Only:             slot.PS4Only,
				HeartCount:          slot.HeartCount,
				ThumbsUp:            slot.ThumbsUp,
				ThumbsDown:          slot.ThumbsDown,
				AverageRating:       slot.AverageRating,
				MMPick:              slot.MMPick,
				ReviewCount:         slot.ReviewCount,
				CommentCount:        slot.CommentCount,
				PhotoCount:          slot.PhotoCount,
				AuthorPhotoCount:    slot.AuthorPhotoCount,
				Tags:                slot.Tags,
				Labels:              slot.Labels,
				FirstPublished:      slot.FirstPublished,
				LastUpdated:         slot.LastUpdated,
				Genre:               slot.Genre,
				CommentsEnabled:     slot.CommentsEnabled,
				ReviewsEnabled:      slot.ReviewsEnabled,
				PublishedIn:         slot.PublishedIn,
				PlayCount:           slot.PlayCount,
				CompletionCount:     slot.CompletionCount,
				LBP1PlayCount:       slot.LBP1PlayCount,
				LBP1CompletionCount: slot.LBP1CompletionCount,
				LBP1UniquePlayCount: slot.LBP1UniquePlayCount,
				LBP2PlayCount:       slot.LBP2PlayCount,
				LBP2CompletionCount: slot.LBP2CompletionCount,
				UniquePlayCount:     slot.UniquePlayCount,
				LBP3PlayCount:       slot.LBP3PlayCount,
				LBP3CompletionCount: slot.LBP3CompletionCount,
				LBP3UniquePlayCount: slot.LBP3UniquePlayCount,
				Got:                 slot.Got,
			}
		}

		dest.Create(slots)
		return nil
	})
	if err != nil {
		panic(err)
	}
	slog.Info("Finished processing slots", "duration", time.Since(startTime), "rowsProcessed", totalRows)
	return time.Since(startTime), totalRows
}

func MigrateUsers(src *gorm.DB, dest *gorm.DB) (time.Duration, int) {
	startTime := time.Now()
	users := make([]model.User, 0, 1000)
	totalRows := 0
	err := gorm.G[model.ImportUser](src).FindInBatches(context.Background(), 1000, func(importUsers []model.ImportUser, batch int) error {
		totalRows += len(importUsers)
		slog.Info("Processing batch of users", "batch", batch, "rows", len(importUsers), "totalRows", totalRows)
		users = make([]model.User, len(importUsers))

		for i, user := range importUsers {
			users[i] = model.User{
				NpHandle:                  user.NpHandle,
				Icon:                      user.Icon,
				Lbp1UsedSlots:             user.Lbp1UsedSlots,
				EntitledSlots:             user.EntitledSlots,
				FreeSlots:                 user.FreeSlots,
				CrossControlUsedSlots:     user.CrossControlUsedSlots,
				CrossControlEntitledSlots: user.CrossControlEntitledSlots,
				CrossControlFreeSlots:     user.CrossControlFreeSlots,
				LBP2UsedSlots:             user.LBP2UsedSlots,
				LBP2EntitledSlots:         user.LBP2EntitledSlots,
				LBP2FreeSlots:             user.LBP2FreeSlots,
				LBP3UsedSlots:             user.LBP3UsedSlots,
				LBP3EntitledSlots:         user.LBP3EntitledSlots,
				LBP3FreeSlots:             user.LBP3FreeSlots,
				HeartCount:                user.HeartCount,
				Planets:                   user.Planets,
				Yay2:                      user.Yay2,
				Boo2:                      user.Boo2,
				LocationX:                 user.LocationX,
				LocationY:                 user.LocationY,
				CommentsEnabled:           user.CommentsEnabled,
				Pins0:                     user.Pins0,
				Pins1:                     user.Pins1,
				Pins2:                     user.Pins2,
			}
		}
		dest.Create(users)
		return nil
	})
	if err != nil {
		panic(err)
	}
	slog.Info("Finished processing users", "duration", time.Since(startTime), "rowsProcessed", totalRows, "err", src.Error)
	return time.Since(startTime), totalRows
}
