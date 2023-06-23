package main

import (
	"bipgit/internal/configs"
	"bipgit/internal/server"
)

// @title Bip Git
// @description Bip Git Backend server.
// @schemes http https
// @termsOfService https://bip.so/terms-of-service/

// @contact.name API Support
// @contact.url https://bip.so
// @contact.email santhosh@bip.so

// @license.name Apache 2.0
// @licence.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apiKey bearerAuth
// @in header
// @name Authorization
// Deploying
func main() {
	configs.Init(".env", ".")
	server.InitServer()
}
