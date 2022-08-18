package cache

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/minio/minio-go"
)

const cacheRoot = ".cache"

type Config struct {
	Debug                  bool     `envconfig:"PLUGIN_DEBUG" default:"false"`
	S3Endpoint             string   `envconfig:"PLUGIN_S3_ENDPOINT"`
	S3AccessKeyID          string   `envconfig:"PLUGIN_S3_ACCESS_KEY_ID"`
	S3Secret               string   `envconfig:"PLUGIN_S3_SECRET"`
	S3Bucket               string   `envconfig:"PLUGIN_S3_BUCKET"`
	RemoteRootDir          string   `envconfig:"PLUGIN_REMOTE_ROOT_DIR"`
	Restore                bool     `envconfig:"PLUGIN_RESTORE" default:"false"`
	Refresh                bool     `envconfig:"PLUGIN_REFRESH" default:"false"`
	CacheableRelativePaths []string `envconfig:"PLUGIN_CACHEABLE_RELATIVE_PATHS" default:""`
	CacheableAbsolutePaths []string `envconfig:"PLUGIN_CACHEABLE_ABSOLUTE_PATHS" default:""`
}

type Cache struct {
	cfg      Config
	s3Client *minio.Client
}

func (c Cache) Execute() error {
	if c.cfg.Refresh {
		err := c.refresh()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	if c.cfg.Restore {
		err := c.restore()
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (c Cache) refresh() error {
	log.Println("Start refreshing cache")
	err := useCacheDir(func() error {
		err := c.archiveAndUpload(c.cfg.CacheableRelativePaths, func(path string) string {
			return path
		})
		if err != nil {
			log.Println(err)
			return err
		}

		return c.archiveAndUpload(c.cfg.CacheableAbsolutePaths, func(path string) string {
			return path[1:]
		})
	})
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Finish refreshing cache")
	return nil
}

func (c Cache) restore() error {
	log.Println("Start restoring cache")
	err := useCacheDir(func() error {
		err := c.downloadAndUnArchive(c.cfg.CacheableRelativePaths, func(path string) string {
			return path
		})
		if err != nil {
			log.Println(err)
			return err
		}

		return c.downloadAndUnArchive(c.cfg.CacheableAbsolutePaths, func(path string) string {
			return path[1:]
		})
	})
	if err != nil {
		log.Println(err)
		return err
	}

	log.Println("Finish restoring cache")
	return nil
}

func (c Cache) archiveAndUpload(paths []string, getPathKey func(path string) string) error {
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
		s3Path := filepath.Join(c.cfg.RemoteRootDir, pathKey)

		log.Printf("Start uploading archive: archiveFilePath=%v\n", fullDestPath)
		_, err = c.s3Client.FPutObject(c.cfg.S3Bucket, s3Path, fullDestPath, minio.PutObjectOptions{})
		if err != nil {
			log.Println(err)
			return err
		}

		log.Printf("Finish uploading archive: archiveFilePath=%v\n", fullDestPath)
	}

	return nil
}

func (c Cache) downloadAndUnArchive(paths []string, getPathKey func(path string) string) error {
	for index, path := range paths {
		fileName := strconv.FormatInt(int64(index), 10) + ".zip"
		fullZipDestPath := filepath.Join(cacheRoot, fileName)
		pathKey := getPathKey(path)
		s3Path := filepath.Join(c.cfg.RemoteRootDir, pathKey)

		log.Printf("Start downloading archive: path=%v, fullZipDestPath=%v\n", path, fullZipDestPath)
		err := c.s3Client.FGetObject(c.cfg.S3Bucket, s3Path, fullZipDestPath, minio.GetObjectOptions{})
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

func New(cfg Config) (Cache, error) {
	if cfg.Debug {
		log.Println(cfg)
	}

	if cfg.Restore && cfg.Refresh {
		return Cache{}, errors.New("restore & refresh are exclusive")
	}

	if !cfg.Restore && !cfg.Refresh {
		return Cache{}, errors.New("plugin must run either in restore or refresh mode")
	}

	s3Client, err := minio.New(cfg.S3Endpoint, cfg.S3AccessKeyID, cfg.S3Secret, true)
	if err != nil {
		log.Println(err)
		return Cache{}, err
	}

	return Cache{
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
