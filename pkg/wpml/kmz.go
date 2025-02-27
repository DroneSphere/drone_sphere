package wpml

import (
	"archive/zip"
	"os"
	"path/filepath"
)

func GenerateKMZ(filename string, template, wayline string) error {
	zipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer func(zipWriter *zip.Writer) {
		err := zipWriter.Close()
		if err != nil {
			panic(err)
		}
	}(zipWriter)

	if err := addToZip(zipWriter, "wpmz/template.kml", template); err != nil {
		return err
	}
	if err := addToZip(zipWriter, "wpmz/waylines.wpml", wayline); err != nil {
		return err
	}
	return nil
}

func addToZip(zipWriter *zip.Writer, filename, content string) error {
	// 确保文件的目录结构存在
	dir := filepath.Dir(filename)
	_, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   dir + "/",
		Method: zip.Store, // 不压缩目录，只是创建目录
	})
	if err != nil {
		return err
	}

	// 创建文件
	w, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   filename,
		Method: zip.Deflate,
	})
	if err != nil {
		return err
	}

	// 写入文件内容
	_, err = w.Write([]byte(content))
	if err != nil {
		return err
	}

	return nil
}
