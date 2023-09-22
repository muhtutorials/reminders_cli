package main

import (
	"flag"
	"fmt"
	"github.com/muhtutorials/reminders_cli/client"
	"os"
)

func main() {
	backendURLFlag := flag.String("backend", "http://localhost:8000", "Backend API URL")
	helpFlag := flag.Bool("help", false, "Display a helpful message")
	flag.Parse()

	s := client.NewSwitch(*backendURLFlag)

	if *helpFlag || len(os.Args) == 1 {
		s.Help()
		return
	}

	err := s.Switch()
	if err != nil {
		fmt.Println("cmd switch error:", err)
		os.Exit(2)
	}
}
