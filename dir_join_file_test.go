package dirjoinfilego_test

import (
	"testing"

	dirjoinfilego "github.com/arisnotargon/dir_join_file_go"
	"github.com/davecgh/go-spew/spew"
)

func TestDirJoin(t *testing.T) {
	prefix := "."
	djf := &dirjoinfilego.DirJionFile{
		TargetDirPath:  prefix + "/target",
		PassWord:       "123456",
		SourceFilePath: prefix + "/animation.gif.mp4",
		OutPutFilePath: prefix + "/out.mp4",
	}
	n, err := djf.Join()

	// n 15403
	spew.Config.Dump(n, err)
}
