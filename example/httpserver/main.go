package main

func main() {
	apiManager := NewApiManager()
	apiManager.Init()
	apiManager.Start()
	apiManager.Run()
	apiManager.Stop()
}
