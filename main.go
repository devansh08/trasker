package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const TASKS_DIR = ".tasks"
const DATE_FMT = "20060102-150405"
const TASK_MD_FILE = "TASK.md"
const DEFAULT_TASK_CONTENTS = `# New Task

- CATEGORY: TODO|FIX|PERF|SPIKE
- STATUS: ACTIVE|COMPLETED|DROPPED

This is a new task.`

func dirExists(dirName string) bool {
	info, err := os.Stat(dirName)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return info.IsDir()
}

func isInitialized() bool {
	if dirExists(".tasks") {
		return true
	}

	fmt.Println("`.tasks` directory not found. Run `init` to setup Trasker in this project.")
	return false
}

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
			case "init":
				if dirExists(TASKS_DIR) {
					fmt.Println("`.tasks` directory already exists. Skipping initialization.")
				} else {
					err := os.Mkdir(TASKS_DIR, 0755)
					if err != nil {
						panic(err)
					}
					fmt.Println("Trasker initialized.")
				}
			case "new":
				if isInitialized() {
					now := time.Now()
					for {
						taskName := now.Format(DATE_FMT)
						if !dirExists(taskName) {
							dirName := fmt.Sprintf("%s/%s", TASKS_DIR, taskName)
							fileName := fmt.Sprintf("%s/%s/%s", TASKS_DIR, taskName, TASK_MD_FILE)

							os.Mkdir(dirName, 0755)
							os.WriteFile(fileName, []byte(DEFAULT_TASK_CONTENTS), 0644)

							editor := "vim"
							for _, str := range os.Environ() {
								if pair := strings.Split(str, "="); pair[0] == "EDITOR" {
									editor = pair[1]
									break
								}
							}

							cmd := exec.Command(editor, fileName)
							cmd.Stdin = os.Stdin
							cmd.Stdout = os.Stdout
							cmd.Stderr = os.Stderr

							err := cmd.Run()
							if err != nil {
								panic(err)
							}

							fmt.Printf("Task `%s` created successfully.\n", taskName)
							break
						}
						now = now.Add(time.Duration(1 * time.Second))
					}
				}
			case "cls":
				fmt.Print("\033[2J\033[H")
			case "help":
				fmt.Println("COMMANDS")
				fmt.Println("  init\t\tinitialize Trasker in current directory")
				fmt.Println("  new\t\tcreate and edit a new task")
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
