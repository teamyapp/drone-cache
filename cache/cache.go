package cache

import (
	"fmt"
	"log"
)

type Mode string

const (
	persistMode  Mode = "persist"
	retrieveMode Mode = "retrieve"
)

func Execute(cfg Config) error {
	if cfg.Debug {
		log.Printf("%#v\n", cfg)
	}

	storage, err := makeStorage(cfg)
	if err != nil {
		return err
	}

	switch Mode(cfg.Mode) {
	case persistMode:
		log.Println("Start persisting cache")
		err = storage.PersistCache()
		if err != nil {
			log.Println(err)
			return err
		}

		log.Println("Finish persisting cache")
	case retrieveMode:
		log.Println("Start retrieving cache")
		err = storage.RetrieveCache()
		if err != nil {
			log.Println(err)
			return err
		}

		log.Println("Finish retrieving cache")
	}

	return nil
}

func makeStorage(cfg Config) (Storage, error) {
	switch storageType(cfg.StorageType) {
	case s3StorageType:
		return newS3Storage(cfg)
	case volumeStorageType:
		return newVolumeStorage(cfg)
	}

	return nil, fmt.Errorf("unsupported storage type: storageType=%v", s3StorageType)
}
