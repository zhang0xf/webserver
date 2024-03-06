package main

func main() {
	watchManager := NewWatchManager()
	watchManager.Init()
	watchManager.Start()
	watchManager.Run()
}
