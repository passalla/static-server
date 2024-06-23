package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Menentukan default values
	defaultPort := "8080"
	defaultStaticDir := "./"
	defaultDebug := false

	// Mengambil argumen dari command line
	port := flag.String("port", defaultPort, "Port to run the server on")
	staticDir := flag.String("static", defaultStaticDir, "Directory for static files")
	isDebug := flag.Bool("debug", defaultDebug, "Is debug mode?")
	flag.Parse()

	if *isDebug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	// Membuat router Gin
	r := gin.Default()

	r.StaticFS("/", http.Dir(*staticDir))

	// Menjalankan server di port yang diberikan
	r.Run(fmt.Sprintf(":%s", *port))
}
