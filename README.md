dir_join_file是一个为隐蔽存储/传输文件而设计的库，可以将目标目录打包压缩解密后拼接在一个指定文件的末尾，并创建一个新文件保存拼接后的结果

建议选择拼接到视频文件的结尾，因为这样原有视频可以正常播放，隐蔽性更强，并且视频文件通常较大，更便于掩护需要被隐藏的目录

todo: 入口文件

使用方法

a go version of [dir_join_file](https://github.com/arisnotargon/dir_join_file)