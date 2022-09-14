package dirjoinfilego_test

import (
	"testing"

	dirjoinfilego "github.com/arisnotargon/dir_join_file_go"
	"github.com/davecgh/go-spew/spew"
)

var globalOffset int64

func TestDirJoin(t *testing.T) {
	prefix := "./"
	djf := &dirjoinfilego.DirJionFile{
		TargetDirPath:  prefix + "test_target",
		PassWord:       "123456",
		SourceFilePath: prefix + "animation.gif.mp4",
		OutPutFilePath: prefix + "out.mp4",
	}
	n, err := djf.Join()
	globalOffset = n

	spew.Config.Dump("n=====>>>>", n, err)
}

func TestRestore(t *testing.T) {
	prefix := "./"
	djf := &dirjoinfilego.DirJionFile{
		PassWord:       "123456",
		SourceFilePath: prefix + "out.mp4",
		OutPutFilePath: prefix + "output",
		FileOffset:     globalOffset, // FileOffset must be set as the same as the first return value of dirjoinfilego.DirJionFile.Join()
	}
	err := djf.Restore()

	spew.Dump("restore error====>>>", err)
}
