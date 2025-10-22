# cmplist - High-Performance File Comparison Tool

A fast Golang tool for comparing two files and analyzing the presence of each item.

## Installation

### Prerequisites
- Go 1.21 or higher

### Build
```bash
go mod tidy
go build -o cmplist cmplist.go
```

## Usage

### Basic Usage
```bash
./cmplist -1 file1.txt -2 file2.txt
```

### Options

| Option | Short | Description | Default |
|--------|-------|-------------|---------|
| `--file1` | `-1` | First file to compare (required) | - |
| `--file2` | `-2` | Second file to compare (required) | - |
| `--color` | - | Enable color output | true |
| `--stats` | - | Show statistics | false |
| `--format` | `-f` | Output format (csv, table, json) | csv |
| `--group` | `-g` | Group output by category | false |
| `--workers` | `-w` | Number of workers for parallel processing | CPU count |
| `--progress` | `-p` | Show progress | false |

### Output Formats

#### 1. CSV Format (Default)
```bash
./cmplist -1 file1.txt -2 file2.txt
```
```
apple,1
banana,2
cherry,3
```

#### 2. Table Format
```bash
./cmplist -1 file1.txt -2 file2.txt --format table
```

#### 3. JSON Format
```bash
./cmplist -1 file1.txt -2 file2.txt --format json
```

#### 4. Grouped Display
```bash
./cmplist -1 file1.txt -2 file2.txt --group --format table
```

### Statistics Display
```bash
./cmplist -1 file1.txt -2 file2.txt --stats
```

### Progress Display
```bash
./cmplist -1 file1.txt -2 file2.txt --progress
```

## Output Description

### Classification Values
- `1`: Items only in FILE1
- `2`: Items in both files
- `3`: Items only in FILE2

### Color Display
- 🔵 **Blue**: Only in FILE1
- 🟢 **Green**: In both files
- 🔴 **Red**: Only in FILE2

## Examples

### Basic Comparison
```bash
./cmplist -1 list1.txt -2 list2.txt
```

### Statistics with Grouped Display
```bash
./cmplist -1 list1.txt -2 list2.txt --stats --group --format table
```

### JSON Format Output
```bash
./cmplist -1 list1.txt -2 list2.txt --format json --stats
```

### Large File Processing
```bash
./cmplist -1 large1.txt -2 large2.txt --progress --workers 8
```
