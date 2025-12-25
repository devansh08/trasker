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

type Category int

const (
	TODO Category = iota
	FIX
	PERF
	SPIKE
	INVALID_CATEGORY
)

func getCategory(category string) Category {
	switch category {
	case "TODO":
		return TODO
	case "FIX":
		return FIX
	case "PERF":
		return PERF
	case "SPIKE":
		return SPIKE
	default:
		return INVALID_CATEGORY
	}
}

func getCategoryString(category Category) string {
	switch category {
	case TODO:
		return "TODO"
	case FIX:
		return "FIX"
	case PERF:
		return "PERF"
	case SPIKE:
		return "SPIKE"
	default:
		return "INVALID_CATEGORY"
	}
}

type Status int

const (
	ACTIVE Status = iota
	COMPLETED
	DROPPED
	INVALID_STATUS
)

func getStatus(status string) Status {
	switch status {
	case "ACTIVE":
		return ACTIVE
	case "COMPLETED":
		return COMPLETED
	case "DROPPED":
		return DROPPED
	default:
		return INVALID_STATUS
	}
}

func getStatusString(status Status) string {
	switch status {
	case ACTIVE:
		return "ACTIVE"
	case COMPLETED:
		return "COMPLETED"
	case DROPPED:
		return "DROPPED"
	default:
		return "INVALID_STATUS"
	}
}

type Task struct {
	id          string
	name        string
	category    Category
	status      Status
	description string
}

var tasks []Task
var categoryTasks map[Category][]*Task
var statusTasks map[Status][]*Task

var categoryMaxLength int
var statusMaxLength int

func dirExists(dirName string) bool {
	info, err := os.Stat(dirName)
	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return info.IsDir()
}

func isInitialized() bool {
	if dirExists(TASKS_DIR) {
		return true
	}

	fmt.Printf("`%s` directory not found. Run `init` to setup Trasker in this project.\n", TASKS_DIR)
	return false
}

func updateTaskFromFile(dirName string) {
	task := Task{id: dirName}

	data, err := os.ReadFile(fmt.Sprintf("%s/%s/%s", TASKS_DIR, dirName, TASK_MD_FILE))
	if err != nil {
		panic(err)
	}

	contents := strings.Split(string(data), "\n")

	task.name = strings.Split(contents[0], "# ")[1]
	task.category = getCategory(strings.Split(contents[2], "- CATEGORY: ")[1])
	task.status = getStatus(strings.Split(contents[3], "- STATUS: ")[1])
	task.description = strings.Trim(strings.Join(contents[5:], "\n"), "\n ")

	if categoryTasks[task.category] == nil {
		categoryTasks[task.category] = []*Task{&task}
	} else {
		categoryTasks[task.category] = append(categoryTasks[task.category], &task)
	}
	if len(getCategoryString(task.category)) > categoryMaxLength {
		categoryMaxLength = len(getCategoryString(task.category))
	}

	if statusTasks[task.status] == nil {
		statusTasks[task.status] = []*Task{&task}
	} else {
		statusTasks[task.status] = append(statusTasks[task.status], &task)
	}
	if len(getStatusString(task.status)) > statusMaxLength {
		statusMaxLength = len(getStatusString(task.status))
	}
}

func loadTasks() {
	if tasks == nil {
		tasks = []Task{}
		categoryTasks = map[Category][]*Task{}
		statusTasks = map[Status][]*Task{}

		categoryMaxLength = 0
		statusMaxLength = 0

		entries, err := os.ReadDir(TASKS_DIR)
		if err != nil {
			panic(err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				updateTaskFromFile(entry.Name())
			}
		}
	}
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
			opts := parts[1:]

			switch cmd {
			case "init":
				if dirExists(TASKS_DIR) {
					fmt.Printf("`%s` directory already exists. Skipping initialization.\n", TASKS_DIR)
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

							updateTaskFromFile(taskName)

							fmt.Printf("Task `%s` created successfully.\n", taskName)
							break
						}
						now = now.Add(time.Duration(1 * time.Second))
					}
				}
			case "ls":
				if isInitialized() {
					loadTasks()

					if tasks != nil {
						if statusMode, categoryMode :=
							len(opts) == 0 || (len(opts) == 1 && opts[0] == "STATUS"),
							len(opts) == 1 && opts[0] == "CATEGORY"; statusMode || categoryMode {
							pad := 1
							if len(tasks) > 10 {
								pad = 2
							}

							count := 1
							if statusMode {
								for status := range INVALID_STATUS {
									fmt.Printf("%s Tasks:\n", getStatusString(status))
									for _, task := range statusTasks[status] {
										fmt.Printf("%*d - %s | %-*s | %s\n", pad, count, task.id, categoryMaxLength, getCategoryString(task.category), task.name)
										count++
									}
								}
							} else if categoryMode {
								for category := range INVALID_CATEGORY {
									fmt.Printf("%s Tasks:\n", getCategoryString(category))
									for _, task := range categoryTasks[category] {
										fmt.Printf("%*d - %s | %-*s | %s\n", pad, count, task.id, statusMaxLength, getStatusString(task.status), task.name)
										count++
									}
								}
							}
						} else if len(opts) == 1 {
							if status := getStatus(opts[0]); status != INVALID_STATUS {
								pad := 1
								if len(statusTasks) > 10 {
									pad = 2
								}

								count := 1
								fmt.Printf("%s Tasks:\n", opts[0])
								for _, task := range statusTasks[status] {
									fmt.Printf("%*d - %s | %-*s | %s\n", pad, count, task.id, categoryMaxLength, getCategoryString(task.category), task.name)
									count++
								}
							} else if category := getCategory(opts[0]); category != INVALID_CATEGORY {
								pad := 1
								if len(categoryTasks) > 10 {
									pad = 2
								}

								count := 1
								fmt.Printf("%s Tasks:\n", opts[0])
								for _, task := range categoryTasks[category] {
									fmt.Printf("%*d - %s | %-*s | %s\n", pad, count, task.id, statusMaxLength, getStatusString(task.status), task.name)
									count++
								}
							} else {
								fmt.Println("Unknown filter for `ls`. Check `help` for correct filters.")
							}
						} else {
							fmt.Println("Unknown filter for `ls`. Check `help` for correct filters.")
						}
					} else {
						panic("Failed to load tasks")
					}
				}
			case "cls":
				fmt.Print("\033[2J\033[H")
			case "help":
				fmt.Println("COMMANDS")
				fmt.Println("  init \t\t\t\tinitialize Trasker in current directory")
				fmt.Println("  new  \t\t\t\tcreate and edit a new task")
				fmt.Println("  ls   \t\t\t\tlist all tasks")
				fmt.Println("    [CATEGORY|STATUS]          \t\tlist tasks grouped by category/status")
				fmt.Println("    [TODO|FIX|PERF|SPIKE]      \t\tlist tasks filtered by given category")
				fmt.Println("    [ACTIVE|COMPLETED|DROPPED] \t\tlist tasks filtered by given status")
				fmt.Println("  cls  \t\t\t\tclear the screen")
				fmt.Println("  help \t\t\t\tdisplay this help")
				fmt.Println("  exit \t\t\t\texit the program")
			case "exit":
				loop = false
			default:
				fmt.Printf("Unknown command: `%s`\n", cmd)
			}
		}
	}
}
