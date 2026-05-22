package main

import (
	"os"

	"github.com/KoreanCode/orange-hyper-run/internal/app"
)

func main() {
	os.Exit(app.Main(os.Args[1:]))
}
