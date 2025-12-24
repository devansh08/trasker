# Trasker

Simple project level file based task tracker.

## Directory Structure

```
.tasks
├── 20251224-121342
│   └── TASK.md
│   └── [other files...]
└── 20251224-174459
    └── TASK.md
```

Each task is tracked as a directory, named with the task creation timestamp, containing a TASK.md file and other optional reference files.

### TASK.md File Structure

```markdown
# Task Name

- CATEGORY: TODO|FIX|PERF|SPIKE
- STATUS: ACTIVE|COMPLETED|DROPPED

Task Description
```

This file **must** follow the below format:
- First line: Task name in Heading 1 style
- Second line: Empty
- Third line: Category of the task in bullet points style; should be one of the ones listed
- Fourth line: Status of the task in bullet points style; should be one of the ones listed
- Fifth line: Empty
- Sixth line and onwards: Task description
