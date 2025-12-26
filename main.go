package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"
)

var Version = "dev"

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
		return ""
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
		return ""
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

var taskList []*Task

var categoryMaxLength int
var statusMaxLength int

func runExtCmd(cmd string, args ...string) {
	extCmd := exec.Command(cmd, args...)
	extCmd.Stdin = os.Stdin
	extCmd.Stdout = os.Stdout
	extCmd.Stderr = os.Stderr

	err := extCmd.Run()
	if err != nil {
		panic(err)
	}
}

func openEditor(fileName string) {
	editor := "vim"
	for _, str := range os.Environ() {
		if pair := strings.Split(str, "="); pair[0] == "EDITOR" {
			editor = pair[1]
			break
		}
	}

	runExtCmd(editor, fileName)
}

func containsTaskPtr(task *Task, taskArr []*Task) (bool, int) {
	if task != nil {
		for i, val := range taskArr {
			if val != nil && task.id == val.id {
				return true, i
			}
		}
	}

	return false, -1
}

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

func updateTaskFromFile(dirName string, task *Task) {
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
		categoryTasks[task.category] = []*Task{task}
	} else if flag, index := containsTaskPtr(task, categoryTasks[task.category]); flag {
		categoryTasks[task.category][index] = task
	} else {
		categoryTasks[task.category] = append(categoryTasks[task.category], task)
	}
	if len(getCategoryString(task.category)) > categoryMaxLength {
		categoryMaxLength = len(getCategoryString(task.category))
	}

	if statusTasks[task.status] == nil {
		statusTasks[task.status] = []*Task{task}
	} else if flag, index := containsTaskPtr(task, statusTasks[task.status]); flag {
		statusTasks[task.status][index] = task
	} else {
		statusTasks[task.status] = append(statusTasks[task.status], task)
	}
	if len(getStatusString(task.status)) > statusMaxLength {
		statusMaxLength = len(getStatusString(task.status))
	}
}

func loadTasks() {
	if tasks == nil {
		tasks = []Task{}
		categoryTasks = make(map[Category][]*Task)
		statusTasks = make(map[Status][]*Task)

		categoryMaxLength = 0
		statusMaxLength = 0

		entries, err := os.ReadDir(TASKS_DIR)
		if err != nil {
			panic(err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				updateTaskFromFile(entry.Name(), &Task{id: entry.Name()})
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
					loadTasks()

					now := time.Now()
					for {
						taskName := now.Format(DATE_FMT)
						if !dirExists(taskName) {
							dirName := fmt.Sprintf("%s/%s", TASKS_DIR, taskName)
							fileName := fmt.Sprintf("%s/%s/%s", TASKS_DIR, taskName, TASK_MD_FILE)

							os.Mkdir(dirName, 0755)
							os.WriteFile(fileName, []byte(DEFAULT_TASK_CONTENTS), 0644)

							openEditor(fileName)

							updateTaskFromFile(taskName, &Task{id: taskName})

							fmt.Printf("Task `%s` created successfully.\n", taskName)
							break
						}
						now = now.Add(time.Duration(1 * time.Second))
					}
				}
			case "edit":
				if isInitialized() {
					if len(opts) == 1 {
						index, err := strconv.Atoi(opts[0])
						if err != nil {
							panic(err)
						}

						if index < 0 || index > len(taskList) {
							fmt.Println("Invalid index provided.")
						} else {
							task := taskList[index-1]
							taskName := task.id

							openEditor(fmt.Sprintf("%s/%s/%s", TASKS_DIR, taskName, TASK_MD_FILE))

							updateTaskFromFile(taskName, task)
							fmt.Printf("Task `%s` updated successfully.\n", taskName)
						}
					} else {
						fmt.Println("Index of task to edit not provided. Check `help` for correct usage.")
					}
				}
			case "rm":
				if isInitialized() {
					if len(opts) == 1 {
						index, err := strconv.Atoi(opts[0])
						if err != nil {
							panic(err)
						}

						if index < 0 || index > len(taskList) {
							fmt.Println("Invalid index provided.")
						} else {
							task := taskList[index-1]
							taskName := task.id

							taskDir := fmt.Sprintf("%s/%s", TASKS_DIR, taskName)
							runExtCmd("rm", "--recursive", "--interactive=once", taskDir)

							if !dirExists(taskDir) {
								tasks = slices.DeleteFunc(tasks, func(ele Task) bool {
									return ele.id == task.id
								})

								flag, ind := containsTaskPtr(task, categoryTasks[task.category])
								if flag {
									categoryTasks[task.category] = slices.Delete(categoryTasks[task.category], ind, ind+1)
								}
								flag, ind = containsTaskPtr(task, statusTasks[task.status])
								if flag {
									statusTasks[task.status] = slices.Delete(statusTasks[task.status], ind, ind+1)
								}
								flag, ind = containsTaskPtr(task, taskList)
								if flag {
									taskList = slices.Delete(taskList, ind, ind+1)
								}

								fmt.Printf("Task `%s` deleted successfully.\n", taskName)
							} else {
								fmt.Println("Skipped deleting task. HINT: Enter `y` on deletion confirmation.")
							}
						}
					} else {
						fmt.Println("Index of task to delete not provided. Check `help` for correct usage.")
					}
				}
			case "cat":
				if isInitialized() {
					if len(opts) == 1 {
						index, err := strconv.Atoi(opts[0])
						if err != nil {
							panic(err)
						}

						if index < 0 || index > len(taskList) {
							fmt.Println("Invalid index provided.")
						} else {
							task := taskList[index-1]
							runExtCmd("cat", fmt.Sprintf("%s/%s/%s", TASKS_DIR, task.id, TASK_MD_FILE))
						}
					} else {
						fmt.Println("Index of task to delete not provided. Check `help` for correct usage.")
					}
				}
			case "ls":
				if isInitialized() {
					loadTasks()

					if tasks != nil {
						taskList = []*Task{}

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
										taskList = append(taskList, task)
										fmt.Printf("%*d - %s | %-*s | %s\n", pad, count, task.id, categoryMaxLength, getCategoryString(task.category), task.name)
										count++
									}
								}
							} else if categoryMode {
								for category := range INVALID_CATEGORY {
									fmt.Printf("%s Tasks:\n", getCategoryString(category))
									for _, task := range categoryTasks[category] {
										taskList = append(taskList, task)
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
								taskList = statusTasks[status]
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
								taskList = categoryTasks[category]
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
				fmt.Println("  init         \t\t\t\t initialize Trasker in current directory")
				fmt.Println("  new          \t\t\t\t create and edit a new task")
				fmt.Println("  ls           \t\t\t\t list all tasks")
				fmt.Println("    [CATEGORY|STATUS]          \t\t\t list tasks grouped by category/status")
				fmt.Println("    [TODO|FIX|PERF|SPIKE]      \t\t\t list tasks filtered by given category")
				fmt.Println("    [ACTIVE|COMPLETED|DROPPED] \t\t\t list tasks filtered by given status")
				fmt.Println("  edit <index> \t\t\t\t edit mentioned task from list (see `ls`)")
				fmt.Println("  rm <index>   \t\t\t\t delete mentioned task from list (see `ls`)")
				fmt.Println("  cat <index>  \t\t\t\t display mentioned task from list (see `ls`)")
				fmt.Println("  cls          \t\t\t\t clear the screen")
				fmt.Println("  help         \t\t\t\t display this help")
				fmt.Println("  version      \t\t\t\t print the version")
				fmt.Println("  exit         \t\t\t\t exit the program")
			case "version":
				fmt.Println(Version)
			case "exit":
				loop = false
			default:
				fmt.Printf("Unknown command: `%s`\n", cmd)
			}
		}
	}
}
