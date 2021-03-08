package realmain

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sanposhiho/gomockhandler/model"
	"github.com/sanposhiho/gomockhandler/realmain/util"
	"golang.org/x/sync/errgroup"
)

func (r Runner) Mockgen() {
	ch, err := r.ChunkRepo.Get(r.Args.ConfigPath)
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	g, _ := errgroup.WithContext(context.Background())
	for _, m := range ch.Mocks {
		g.Go(func() error {
			var destination string
			switch m.Mode {
			case model.Unknown:
				log.Printf("unknown mock detected\n")
				return nil
			case model.ReflectMode:
				err = m.ReflectModeRunner.Run()
				destination = m.ReflectModeRunner.Destination
			case model.SourceMode:
				err = m.SourceModeRunner.Run()
				destination = m.SourceModeRunner.Destination
			}
			if err != nil {
				return fmt.Errorf("run mockgen: %v", err)
			}

			checksum, err := util.MockChackSum(destination)
			if err != nil {
				return fmt.Errorf("calculate checksum of the mock: %v", err)
			}

			m.CheckSum = checksum
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		log.Fatalf("failed to run: %v", err.Error())
	}

	if err := r.ChunkRepo.Put(ch, r.Args.ConfigPath); err != nil {
		log.Fatalf("failed to put config: %v", err)
	}
	return
}

func pathInProject(projectRoot, path string) string {
	return strings.Replace(path, projectRoot, ".", 1)
}