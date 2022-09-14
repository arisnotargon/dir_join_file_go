// @Description append a folder to the end of a specified file,with gzip compressed and aes crypto
// @Auth https://github.com/arisnotargon
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
	"strings"
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

	FileOffset int64
}

// return n int the length of original source file,err error
func (d *DirJionFile) Join() (n int64, err error) {
	// check target directoy exist
	stat, err := os.Stat(d.TargetDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(d.TargetDirPath, os.ModePerm)
			if err != nil {
				return
			}
		} else if os.IsPermission(err) {
			err = errors.New("target path permision error")
			return
		} else {
			return
		}
	}

	if !stat.IsDir() {
		err = errors.New("target path is not a directory")
		return
	}

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

	_, err = outFile.Write(p)
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

	n = int64(len(data))
	// d.buf.WriteTo(outFile)
	outFile.Write(data)

	return
}

func (d *DirJionFile) Restore() (err error) {
	// check output path exist
	outPathStat, err := os.Stat(d.OutPutFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(d.OutPutFilePath, os.ModePerm)
			if err != nil {
				return
			}
			outPathStat, err = os.Stat(d.OutPutFilePath)
		} else if os.IsPermission(err) {
			err = errors.New("out put path permision error")
			return
		} else {
			return
		}
	}

	if !outPathStat.IsDir() {
		err = errors.New("out put path is not a directory")
		return
	}

	// read joined file with offset
	sourceFile, err := os.Open(d.SourceFilePath)
	if err != nil {
		return
	}
	defer sourceFile.Close()

	sourceFile.Seek(-d.FileOffset, io.SeekEnd)

	sourceData, err := io.ReadAll(sourceFile)

	// decrypto
	originData, err := d.bufDecrypto(&sourceData)
	if err != nil {
		return
	}

	sourceBufer := bytes.NewBuffer(originData)

	gzipReader, err := gzip.NewReader(sourceBufer)
	if err != nil {
		return
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)

Loop:
	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			break Loop
		case err != nil:
			return err
		case header == nil:
			continue
		}
		targetPath := filepath.Join(d.OutPutFilePath, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			info, err := os.Stat(targetPath)
			dirExist := (err == nil || os.IsExist(err)) && info.IsDir()
			if !dirExist {
				err = os.MkdirAll(targetPath, os.ModePerm)
				if err != nil {
					return err
				}
			}

		case tar.TypeReg:
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			if err2 := file.Close(); err2 != nil {
				return err2
			}
			if err != nil {
				return err
			}
		}
	}

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
		err1 := fileReader.Close()
		if err1 != nil {
			reErr = err1
		}
	}(fileReader)

	if reErr != nil {
		return
	}

	_, reErr = io.Copy(d.tarWriter, fileReader)
	if reErr != nil {
		return
	}

	return
}

// bufCrypto aes crypto
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

func (d *DirJionFile) bufDecrypto(ciphertext *[]byte) (reData []byte, err error) {
	d.initPassWord()
	err = d.initIv(aes.BlockSize)
	if err != nil {
		return
	}
	block, err := aes.NewCipher(d.passWordBytes)
	if err != nil {
		return
	}
	blockMode := cipher.NewCBCDecrypter(block, d.Iv)
	reData = make([]byte, len(*ciphertext))
	blockMode.CryptBlocks(reData, *ciphertext)
	reData = pkcs7UnPadding(reData)

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

func pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
