package dirjoinfilego

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"errors"
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

	passWordBytes []byte

	// path to the source file which will be appended to
	SourceFilePath string

	// path to the out put file
	OutPutFilePath string

	tarWriter *tar.Writer

	buf *bytes.Buffer

	// iv for aes crypto
	Iv []byte
}

// return n int the length of original source file,err error
func (d *DirJionFile) Join() (n int, err error) {
	spew.Dump("start==============")
	sourceReader, err := os.Open(d.SourceFilePath)

	if err != nil {
		return
	}

	// get data from source file
	p, err := io.ReadAll(sourceReader)

	if err != nil {
		return
	}

	// 测试写入文件
	outFile, _ := os.Create(d.OutPutFilePath)
	defer outFile.Close()

	n, err = outFile.Write(p)
	if err != nil {
		return
	}

	d.buf = &bytes.Buffer{}
	gzipWriter := gzip.NewWriter(d.buf)
	defer gzipWriter.Close()

	d.tarWriter = tar.NewWriter(gzipWriter)
	defer d.tarWriter.Close()

	// handle the specified directory recursively by filepath.Walk from go standard library
	err = filepath.Walk(d.TargetDirPath, d.doTar)

	// must close writers manually before buf.WriteTo(outFile)
	d.tarWriter.Close()
	gzipWriter.Close()

	// crypto
	data, err := d.bufCrypto()
	spew.Dump("data,err := d.bufCrypto()===>", data, err)

	// d.buf.WriteTo(outFile)
	outFile.Write(data)

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

	reErr = d.tarWriter.WriteHeader(fileInfoHeader)
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

	n, reErr := io.Copy(d.tarWriter, fileReader)
	if reErr != nil {
		return
	}

	spew.Dump(fileName + "====> has been written to tar,len" + strconv.Itoa(int(n)))

	return
}

func (d *DirJionFile) bufCrypto() (reData []byte, err error) {
	d.initPassWord()
	err = d.initIv(aes.BlockSize)
	if err != nil {
		return
	}
	block, err := aes.NewCipher(d.passWordBytes)
	if err != nil {
		return
	}
	blockSize := block.BlockSize()

	plaintext := pkcs7Padding(d.buf.Bytes(), blockSize)
	blockMode := cipher.NewCBCEncrypter(block, d.Iv)

	reData = make([]byte, len(plaintext))

	blockMode.CryptBlocks(reData, plaintext)

	return
}

// Initialize password for 32 length
func (d *DirJionFile) initPassWord() (err error) {
	tempPw := make([]byte, 32)
	pdBytes := []byte(d.PassWord)
	for idx := range tempPw {
		if idx < len(pdBytes) {
			tempPw[idx] = pdBytes[idx]
		} else {
			tempPw[idx] = byte(0)
		}
	}

	d.passWordBytes = tempPw

	return
}

func (d *DirJionFile) initIv(blockSize int) (err error) {
	if len(d.Iv) == 0 {
		tempIv := d.passWordBytes
		if len(tempIv) < blockSize {
			for i := 0; i < (blockSize - len(tempIv)); i++ {
				tempIv = append(tempIv, '0')
			}
		}

		d.Iv = tempIv[:blockSize]
	}

	if len(d.Iv) != blockSize {
		return errors.New("iv length error")
	}
	return
}

func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
