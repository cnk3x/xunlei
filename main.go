package main

func main() {
	RunCommand(
		&Command{Name: "run", Run: run, Desc: "启动迅雷"},
		&Command{Name: "install", Run: install, Desc: "安装"},
		&Command{Name: "clean", Run: clean, Desc: "清理(仅清理几个程序自动添加到系统的文件)"},
	)
}
