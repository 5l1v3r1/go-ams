package amsutil

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/recruit-tech/go-ams"
)

func WaitJob(ctx context.Context, client *ams.Client, jobID string, duration time.Duration) error {
	for {
		current, err := client.GetJob(ctx, jobID)
		if err != nil {
			return err
		}

		if current.State == ams.JobError {
			return errors.New("job failed")
		}
		if current.State == ams.JobCanceled {
			return errors.New("job canceled")
		}
		if current.State == ams.JobFinished {
			return nil
		}
		time.Sleep(duration)
	}
}
