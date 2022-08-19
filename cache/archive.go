package cache

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func archive(srcPath string, destPath string) error {
	f, err := os.Create(destPath)
	if err != nil {
		log.Println(err)
		return err
	}

	defer f.Close()

	zipWriter := zip.NewWriter(f)
	defer zipWriter.Close()

	return filepath.Walk(srcPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Println(err)
			return err
		}

		if isSymLink(info.Mode()) {
			link, err := os.Readlink(path)
			if err != nil {
				log.Println(err)
				return err
			}

			header.Extra = []byte(link)
		}

		header.Method = zip.Deflate
		header.Name, err = filepath.Rel(srcPath, path)
		if err != nil {
			log.Println(err)
			return err
		}

		if info.IsDir() {
			header.Name += "/"
		}

		fileWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			log.Println(err)
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		srcFile, err := os.Open(path)
		if err != nil {
			log.Println(err)
			return err
		}

		defer srcFile.Close()

		_, err = io.Copy(fileWriter, srcFile)
		return err
	})
}

func unArchive(srcPath string, destPath string) error {
	reader, err := zip.OpenReader(srcPath)
	if err != nil {
		log.Println(err)
		return err
	}

	defer reader.Close()
	for _, archivedFile := range reader.File {
		fullPath := filepath.Join(destPath, archivedFile.Name)
		info := archivedFile.FileInfo()
		if info.IsDir() {
			err = os.MkdirAll(fullPath, info.Mode())
			if err != nil {
				log.Println(err)
				return err
			}

			continue
		}

		err = os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
		if isSymLink(info.Mode()) {
			link := string(archivedFile.Extra)
			err = os.Link(link, fullPath)
			if err != nil {
				log.Println(err)
				return err
			}

			continue
		}

		originalFile, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, info.Mode())
		if err != nil {
			log.Println(err)
			return err
		}

		archivedFileReader, err := archivedFile.Open()
		if err != nil {
			log.Println(err)
			return err
		}

		_, err = io.Copy(originalFile, archivedFileReader)
		if err != nil {
			log.Println(err)
			return err
		}

		originalFile.Close()
		archivedFileReader.Close()
	}

	return nil
}

func isSymLink(mode fs.FileMode) bool {
	return mode&os.ModeSymlink != 0
}
