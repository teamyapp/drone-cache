package cache

import (
	"archive/zip"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

const originalFilePerm = 777

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

		if info.IsDir() {
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
			err = os.MkdirAll(fullPath, os.ModePerm)
			if err != nil {
				log.Println(err)
				return err
			}

			continue
		}

		err = os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
		originalFile, err := os.OpenFile(fullPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, originalFilePerm)
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
