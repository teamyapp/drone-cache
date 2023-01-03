package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"
)

const relativePrefix = "relative"
const absolutePrefix = "absolute"
const completeMetaFileName = ".complete"

type VolumeStorage struct {
	cfg Config
}

var _ Storage = (*VolumeStorage)(nil)

func (v VolumeStorage) PersistCache() error {
	persistDir, err := v.cachePersistDir()
	if err != nil {
		log.Println(err)
		return err
	}

	// make sure .complete exist in case previous persist was interrupted
	cacheCompleteFile := filepath.Join(persistDir, completeMetaFileName)
	isExist, err := exist(cacheCompleteFile)
	if err != nil {
		log.Println(err)
		return err
	}

	if isExist {
		log.Println("Cache found, skip persisting")
		return nil
	}

	for _, relativePath := range v.cfg.CacheableRelativePaths {
		cachePath := filepath.Join(persistDir, relativePrefix, relativePath)
		log.Printf("[RelativePath] copying: cachePath=%v originalPath=%v\n", cachePath, relativePath)
		err = copyRec(relativePath, cachePath)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	for _, absolutePath := range v.cfg.CacheableAbsolutePaths {
		cachePath := filepath.Join(persistDir, absolutePrefix, absolutePath[1:])
		log.Printf("[AbsolutePath] copying files: cachePath=%v originalPath=%v\n", cachePath, absolutePath)
		err = copyRec(absolutePath, cachePath)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	file, err := os.Create(filepath.Join(persistDir, completeMetaFileName))
	defer file.Close()
	log.Println("created .complete file")
	return err
}

func (v VolumeStorage) RetrieveCache() error {
	persistDir, err := v.cachePersistDir()
	if err != nil {
		log.Println(err)
		return err
	}

	isExist, err := exist(persistDir)
	if err != nil {
		log.Println(err)
		return err
	}

	if !isExist {
		log.Printf("Cache not found, skip retrieval: persistDir=%v\n", persistDir)
		return nil
	}

	log.Println("Cache found")

	cacheCompleteFile := filepath.Join(persistDir, completeMetaFileName)
	isExist, err = exist(cacheCompleteFile)
	if err != nil {
		log.Println(err)
		return err
	}

	if !isExist {
		log.Println("Cache incomplete, skip retrieval")
		return nil
	}

	for _, relativePath := range v.cfg.CacheableRelativePaths {
		relativeDir := filepath.Dir(relativePath)
		err = os.MkdirAll(relativeDir, 0700)
		if err != nil {
			log.Println(err)
			return err
		}

		cachePath := filepath.Join(persistDir, relativePrefix, relativePath)
		isExist, err = exist(cachePath)
		if err != nil {
			log.Println(err)
			return err
		}

		if !isExist {
			err = fmt.Errorf("[Relative] target path not found: cachePath=%v", cachePath)
			log.Println(err)
			return err
		}

		os.RemoveAll(relativePath)
		err = copyRec(cachePath, relativePath)
		log.Printf("[RelativePath] copying files: cachePath=%v originalPath=%v\n", cachePath, relativePath)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	absPersistDir, err := filepath.Abs(persistDir)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, absolutePath := range v.cfg.CacheableAbsolutePaths {
		absoluteDir := filepath.Dir(absolutePath)
		err = os.MkdirAll(absoluteDir, 0700)
		if err != nil {
			log.Println(err)
			return err
		}

		cachePath := filepath.Join(absPersistDir, absolutePrefix, absolutePath[1:])
		isExist, err = exist(cachePath)
		if err != nil {
			log.Println(err)
			return err
		}

		if !isExist {
			err = fmt.Errorf("[Absolute] target path not found: cachePath=%v", cachePath)
			log.Println(err)
			return err
		}

		os.RemoveAll(absolutePath)
		err = copyRec(cachePath, absolutePath)
		fmt.Printf("[AbsolutePath] copying files: cachePath=%v originalPath=%v\n", cachePath, absolutePath)
		if err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}

func (v VolumeStorage) cachePersistDir() (string, error) {
	cachePersistDir := filepath.Join(v.cfg.VolumeCacheRootDir, v.cfg.RepoName)
	if len(v.cfg.VersionFilePath) > 0 {
		file, err := os.Open(v.cfg.VersionFilePath)
		if err != nil {
			log.Println(err)
			return "", err
		}

		defer file.Close()

		hasher := sha256.New()
		_, err = io.Copy(hasher, file)
		if err != nil {
			log.Println(err)
			return "", err
		}

		hash := hex.EncodeToString(hasher.Sum(nil))
		cachePersistDir = filepath.Join(cachePersistDir, hash)
	}

	return cachePersistDir, nil
}

func newVolumeStorage(cfg Config) (VolumeStorage, error) {
	return VolumeStorage{
		cfg: cfg,
	}, nil
}

func copyRec(srcPath string, destPath string) error {
	os.RemoveAll(destPath)
	fileInfo, err := os.Lstat(srcPath)
	if err != nil {
		log.Println(err)
		return err
	}

	mode := fileInfo.Mode()
	switch mode & os.ModeType {
	case os.ModeDir:
		err = os.MkdirAll(destPath, mode&os.ModePerm)
		if err != nil {
			log.Println(err)
			return err
		}

		entries, err := os.ReadDir(srcPath)
		if err != nil {
			log.Println(err)
			return err
		}

		for _, entry := range entries {
			entrySrcPath := filepath.Join(srcPath, entry.Name())
			entryDestPath := filepath.Join(destPath, entry.Name())
			err = copyRec(entrySrcPath, entryDestPath)
			if err != nil {
				log.Println(err)
				return err
			}
		}
	case os.ModeSymlink:
		link, err := os.Readlink(srcPath)
		if err != nil {
			log.Println(err)
			return err
		}

		os.RemoveAll(destPath)
		err = os.Symlink(link, destPath)
		if err != nil {
			log.Println(err)
			return err
		}
	default:
		err = copyFile(srcPath, destPath)
		if err != nil {
			log.Println(err)
			return err
		}

		err = os.Chmod(destPath, fileInfo.Mode())
		if err != nil {
			log.Println(err)
			return err
		}
	}

	stat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("fail to get raw syscall.Stat_t for %v", srcPath)
	}

	return os.Lchown(destPath, int(stat.Uid), int(stat.Gid))
}

func copyFile(srcFilePath string, destFilePath string) error {
	outputFile, err := os.Create(destFilePath)
	if err != nil {
		return err
	}

	defer outputFile.Close()

	inputFile, err := os.Open(srcFilePath)
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		log.Println(err)
	}

	return err
}

func exist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		log.Println(err)
		return false, err
	}

	return true, nil
}
