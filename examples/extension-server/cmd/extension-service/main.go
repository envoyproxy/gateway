package main

import "github.com/exampleorg/envoygateway-extension/internal/service"

func main() {
	service.NewExtensionService().Start()
}
