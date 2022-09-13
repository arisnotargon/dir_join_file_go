package dirjoinfilego

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// pack up a directory as a tar buffer,then crypto this buffer and append it to the end of a specified file
type DirJionFile struct {
	// path to the target directory
	TargetDirPath string

	// key
	PassWord string

	// path to the source file which will be appended to
	SourceFilePath string

	// path to the out put file
	OutPutFilePath string

	gzipWriter *gzip.Writer

	tarWrite *tar.Writer

	outFile *os.File
}

// return n int the length of original source file,err error
func (d *DirJionFile) Join() (n int, err error) {
	sourceReader, err := os.Open(d.SourceFilePath)

	if err != nil {
		return
	}

	p, err := io.ReadAll(sourceReader)

	if err != nil {
		return
	}
	_ = p

	// 测试写入文件
	d.outFile, _ = os.Create(d.OutPutFilePath)
	defer d.outFile.Close()

	// n, err = outWriter.Write(p)
	// if err != nil {
	// 	return
	// }

	// d.gzipWriter = gzip.NewWriter(buf)
	// defer d.gzipWriter.Close()

	d.tarWrite = tar.NewWriter(d.outFile)
	defer d.tarWrite.Close()

	// handle the specified directory recursively by filepath.Walk from go standard library
	err = filepath.Walk(d.TargetDirPath, d.doTar)

	spew.Dump("filepath.Walk err=====>", err)

	return
}

func (d *DirJionFile) doTar(fileName string, fileInfo os.FileInfo, err error) (reErr error) {
	if err != nil {
		reErr = err
		return
	}

	fileInfoHeader, reErr := tar.FileInfoHeader(fileInfo, "")
	if reErr != nil {
		return
	}

	// reset the fileInfoHeader.Name to the fileName with path,remove the first PathSeparator if the path is absolute path
	fileInfoHeader.Name = strings.TrimPrefix(fileName, string(os.PathSeparator))

	reErr = d.tarWrite.WriteHeader(fileInfoHeader)
	if reErr != nil {
		return
	}

	if !fileInfo.Mode().IsRegular() {
		// skip
		return nil
	}

	fileReader, reErr := os.Open(fileName)
	defer func(fileReader *os.File) {
		err := fileReader.Close()
		spew.Dump("fileReader close err===>", err)
	}(fileReader)

	if reErr != nil {
		return
	}

	n, reErr := io.Copy(d.tarWrite, fileReader)
	if reErr != nil {
		return
	}

	spew.Dump(fileName + "====> has been written to tar,len" + strconv.Itoa(int(n)))

	return
}
