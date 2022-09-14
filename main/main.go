package main

import (
	"flag"
	"fmt"

	dirjoinfilego "github.com/arisnotargon/dir_join_file_go"
)

const (
	modeJoin    string = "join"
	modeRestore string = "restore"
)

func main() {
	mode := flag.String("m", "", "-m <mode> {join,restore}, join for pack up a directroy and append it to the end of a specified file,restore for restore the directroy from a file to a specified path")

	sourceFilePath := flag.String("s", "", "<source file path> in join mode,it's the path of the specified file, in restore mode, it's the path of the output file made by join mode")

	password := flag.String("p", "", "[password]")

	outputPath := flag.String("o", "", "<output path> in join mode,the joined file while be save to this path,it must be not exist in join mode;in restore mode,the original directory will be unpack to this path,it must be a directory or not exist in join mode")

	targetDirPath := flag.String("t", "", "-t [target directory] the path of the target directory,this parameter is necessary in join mode")

	offset := flag.Int64("n", 0, "[offset] the offset of joined file,it will be provided after join mode be run ")

	flag.Parse()

	if len(*sourceFilePath) < 1 {
		fmt.Println("source file path is necessary!")
		showHelp()
		return
	}

	if len(*outputPath) < 1 {
		fmt.Println("output path path is necessary!")
		showHelp()
		return
	}

	switch *mode {
	case modeJoin:
		if len(*targetDirPath) < 1 {
			fmt.Println("target directory path is necessary!")
			showHelp()
			return
		}
		djf := &dirjoinfilego.DirJionFile{
			TargetDirPath:  *targetDirPath,
			PassWord:       *password,
			SourceFilePath: *sourceFilePath,
			OutPutFilePath: *outputPath,
		}

		n, err := djf.Join()
		if err != nil {
			fmt.Println("join failed")
			fmt.Println(err.Error())
			return
		}

		fmt.Println("join done! offset:")
		fmt.Println(n)
		return

	case modeRestore:
		if *offset < 1 {
			fmt.Println("offset is necessary!")
			showHelp()
			return
		}

		djf := &dirjoinfilego.DirJionFile{
			PassWord:       *password,
			SourceFilePath: *sourceFilePath,
			OutPutFilePath: *outputPath,
			FileOffset:     *offset, // FileOffset must be set as the same as the first return value of dirjoinfilego.DirJionFile.Join()
		}
		err := djf.Restore()
		if err != nil {
			fmt.Println("restore failed")
			fmt.Println(err.Error())
			return
		}
		fmt.Println("restore done!")
		return

	default:
		fmt.Println("mode error!")
		showHelp()
		return
	}
}

func showHelp() {
	help := `
usage: 
	-m <mode> {join,restore}, join for pack up a directroy and append it to the end of a specified file,restore for restore the directroy from a file to a specified path
	-s <source file path> in join mode,it's the path of the specified file, in restore mode, it's the path of the output file made by join mode
	-p [password]
	-o <output path> in join mode,the joined file while be save to this path,it must be not exist in join mode;in restore mode,the original directory will be unpack to this path,it must be a directory or not exist in join mode
	-t [target directory] the path of the target directory,this parameter is necessary in join mode
	-n [offset] the offset of joined file,it will be provided after join mode be run 
eg:
  join mode:
	dir_join_file -m=join -s=./animation.gif.mp4 -p=123456 -o=out.mp4 -t=./test_target
  restore mode:
    dir_join_file -m=restore -s=out.mp4 -o=./output -n=352 -p=123456
	`

	fmt.Println(help)
}
