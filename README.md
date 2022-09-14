dir_join_file是一个为隐蔽存储/传输文件而设计的库，可以将目标目录打包压缩解密后拼接在一个指定文件的末尾，并创建一个新文件保存拼接后的结果

建议选择拼接到视频文件的结尾，因为这样原有视频可以正常播放，隐蔽性更强，并且视频文件通常较大，更便于掩护需要被隐藏的目录

todo: 入口文件

## 使用方法

1. 引入
使用go get安装
``` bash
go get -u github.com/arisnotargon/dir_join_file_go
```

然后在代码中引入库
```go
    import (
	 dirjoinfilego "github.com/arisnotargon/dir_join_file_go"
    )
```


2. 创建对象实例，运行方法
    1. 打包实例
    ``` go
    djf := &dirjoinfilego.DirJionFile{
		TargetDirPath:  "需要被打包的目录路径",
		PassWord:       "密码",
		SourceFilePath: "载体文件路径",
		OutPutFilePath: "输出文件路径",
	}
    n, err := djf.Join() // 返回值中的n为需要隐藏的目录打包压缩后的长度，解包时需要用到该值
    ```
    2. 解包实例
    ``` go
    djf := &dirjoinfilego.DirJionFile{
		PassWord:       "密码",
		SourceFilePath: "待解包的源文件",
		OutPutFilePath: "保存解包后目录的路径",
		FileOffset:     offset, // 偏移量，该值必须是打包方法的第一个返回值
	}
    err := djf.Restore()
    ```


a go version of [dir_join_file](https://github.com/arisnotargon/dir_join_file)