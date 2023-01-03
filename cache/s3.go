package cache

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/minio/minio-go"
)

const cacheRoot = ".cache"

type S3Storage struct {
	cfg      Config
	s3Client *minio.Client
}

var _ Storage = (*S3Storage)(nil)

func (s S3Storage) PersistCache() error {
	err := useCacheDir(func() error {
		err := s.archiveAndUpload(s.cfg.CacheableRelativePaths, func(path string) string {
			return filepath.Join("relative", path)
		})
		if err != nil {
			log.Println(err)
			return err
		}

		return s.archiveAndUpload(s.cfg.CacheableAbsolutePaths, func(path string) string {
			return filepath.Join("absolute", path[1:])
		})
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s S3Storage) RetrieveCache() error {
	err := useCacheDir(func() error {
		err := s.downloadAndUnArchive(s.cfg.CacheableRelativePaths, func(path string) string {
			return filepath.Join("relative", path)
		})
		if err != nil {
			log.Println(err)
			return err
		}

		return s.downloadAndUnArchive(s.cfg.CacheableAbsolutePaths, func(path string) string {
			return filepath.Join("absolute", path[1:])
		})
	})
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (s S3Storage) archiveAndUpload(paths []string, getPathKey func(path string) string) error {
	for index, path := range paths {
		fileName := strconv.FormatInt(int64(index), 10) + ".zip"
		fullDestPath := filepath.Join(cacheRoot, fileName)

		log.Printf("Start archiving path: path=%v, archiveFilePath=%v\n", path, fullDestPath)
		err := archive(path, fullDestPath)
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("Finish archiving path: path=%v\n", path)
		pathKey := getPathKey(path)
		s3Path := filepath.Join(s.cfg.S3CacheRootDir, pathKey)

		log.Printf("Start uploading archive: archiveFilePath=%v\n", fullDestPath)
		_, err = s.s3Client.FPutObject(s.cfg.S3Bucket, s3Path, fullDestPath, minio.PutObjectOptions{})
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("Finish uploading archive: archiveFilePath=%v\n", fullDestPath)
	}

	return nil
}

func (s S3Storage) downloadAndUnArchive(paths []string, getPathKey func(path string) string) error {
	for index, path := range paths {
		fileName := strconv.FormatInt(int64(index), 10) + ".zip"
		fullZipDestPath := filepath.Join(cacheRoot, fileName)
		pathKey := getPathKey(path)
		s3Path := filepath.Join(s.cfg.S3CacheRootDir, pathKey)

		log.Printf("Start downloading archive: path=%v, fullZipDestPath=%v\n", path, fullZipDestPath)
		err := s.s3Client.FGetObject(s.cfg.S3Bucket, s3Path, fullZipDestPath, minio.GetObjectOptions{})
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("Finish downloading archive: path=%v, fullZipDestPath=%v\n", path, fullZipDestPath)
		log.Printf("Start unArchiving path: fullZipDestPath=%v\n", fullZipDestPath)
		err = unArchive(fullZipDestPath, path)
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("Finish unArchiving path: fullZipDestPath=%v\n", fullZipDestPath)
	}

	return nil
}

func newS3Storage(cfg Config) (S3Storage, error) {
	s3Client, err := minio.New(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3Secret, true)
	if err != nil {
		log.Println(err)
		return S3Storage{}, err
	}

	return S3Storage{
		cfg:      cfg,
		s3Client: s3Client,
	}, nil
}

func useCacheDir(execute func() error) error {
	err := os.RemoveAll(cacheRoot)
	if err != nil {
		log.Println(err)
		return err
	}

	err = os.MkdirAll(cacheRoot, os.ModePerm)
	if err != nil {
		log.Println(err)
		return err
	}

	err = execute()
	if err != nil {
		return err
	}

	return os.RemoveAll(cacheRoot)
}
