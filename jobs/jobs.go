package jobs

import (
	"context"
	"io/fs"
	"log"
	"os"
	"time"

	"github.com/CPU-commits/USACH.dev-Server/db"
	"github.com/CPU-commits/USACH.dev-Server/utils"
	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/semaphore"
)

// Serivce
type JobService struct {
	Scheduler *cron.Cron
}

func (j *JobService) NewJob(spec string, cmd func()) {
	_, err := j.Scheduler.AddFunc(spec, cmd)
	if err != nil {
		panic(err)
	}
	j.Scheduler.Start()
}

func Init() {
	jobService := NewJobService()
	// Init jobs
	// Delete expired tokens
	jobService.NewJob("1 * * * *", func() {
		now := time.Now()
		_, err := usersTokenModel.Use().DeleteMany(db.Ctx, bson.D{{
			Key: "finish_date",
			Value: bson.M{
				"$lte": primitive.NewDateTimeFromTime(now),
			},
		}})
		if err != nil {
			log.Printf("No se completó exitosamente el job Delete expired tokens")
		}
	})
	// Delete unref files
	jobService.NewJob("0 3 * * *", func() {
		entries, err := os.ReadDir(settingsData.MEDIA_FOLDER)
		if err != nil {
			log.Println("1. No se completó exitosamente el job Delete unref files")
		}
		// Delete
		var sem = semaphore.NewWeighted(int64(10))
		ctx := context.Background()

		for _, file := range entries {
			if err := sem.Acquire(ctx, 1); err != nil {
				log.Println("2. No se completó exitosamente el job Delete unref files")
				break
			}
			go func(file fs.DirEntry) {
				if !file.IsDir() {
					hasRef, err := systemFileService.FileHasRef(file.Name())
					if err != nil {
						log.Println("3. No se completó exitosamente el job Delete unref files")
						return
					}
					if !hasRef {
						utils.DeleteFile(file.Name())
					}
				}

				sem.Release(1)
			}(file)
		}
	})
}

func NewJobService() *JobService {
	return &JobService{
		Scheduler: cron.New(
			cron.WithLogger(cron.DefaultLogger),
		),
	}
}
