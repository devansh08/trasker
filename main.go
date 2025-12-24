package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	loop := true
	for loop {
		fmt.Print("> ")

		str, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		if str = strings.TrimSpace(strings.Trim(str, "\n")); str != "" {
			parts := strings.Split(str, " ")
			cmd := parts[0]

			switch cmd {
			case "cls":
				fmt.Print("\033[2J\033[H")
			case "help":
				fmt.Println("COMMANDS")
				fmt.Println("  cls\t\tclear the screen")
				fmt.Println("  help\t\tdisplay this help")
				fmt.Println("  exit\t\texit the program")
			case "exit":
				loop = false
			default:
				fmt.Printf("Unknown command: `%s`\n", cmd)
			}
		}
	}
}
