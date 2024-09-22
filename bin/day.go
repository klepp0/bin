package main

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "regexp"
    "sort"
    "time"
)

func main() {
    // Get the OBSIDIAN_PATH from environment variables
    vaultPath := os.Getenv("OBSIDIAN_PATH")
    if vaultPath == "" {
        fmt.Println("Please set the OBSIDIAN_PATH environment variable.")
        os.Exit(1)
    }

    // Get today's date in YYYY-MM-DD format
    today := time.Now().Format("2006-01-02")

    // Find the directory where daily notes are stored
    dailyNoteDir, err := findDailyNotesDirectory(vaultPath)
    if err != nil {
        fmt.Printf("Error finding daily notes directory: %v\n", err)
        os.Exit(1)
    }

    // If no daily notes directory is found, default to the root of the vault
    if dailyNoteDir == "" {
        dailyNoteDir = vaultPath
    }

    // Construct the file path
    noteName := today + ".md"
    notePath := filepath.Join(dailyNoteDir, noteName)

    // Check if the file exists
    if _, err := os.Stat(notePath); os.IsNotExist(err) {
        // File does not exist, create it
        file, err := os.Create(notePath)
        if err != nil {
            fmt.Printf("Error creating note: %v\n", err)
            os.Exit(1)
        }
        defer file.Close()

        // Add default content to the new note
        file.WriteString("# " + today + "\n\n")
    }

    // Open the note in nvim
    cmd := exec.Command("nvim", notePath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err = cmd.Run()
    if err != nil {
        fmt.Printf("Error opening note in nvim: %v\n", err)
        os.Exit(1)
    }
}

// Function to find the directory containing daily notes
func findDailyNotesDirectory(vaultPath string) (string, error) {
    // Regular expression to match daily note filenames, e.g., "2023-10-06.md"
    dailyNotePattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\.md$`)

    directoryCount := make(map[string]int)

    err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        if dailyNotePattern.MatchString(info.Name()) {
            dir := filepath.Dir(path)
            directoryCount[dir]++
        }

        return nil
    })

    if err != nil {
        return "", err
    }

    // No daily notes found
    if len(directoryCount) == 0 {
        return "", nil
    }

    // Find the directory with the most daily notes
    type dirCount struct {
        Dir   string
        Count int
    }

    var dirs []dirCount
    for dir, count := range directoryCount {
        dirs = append(dirs, dirCount{dir, count})
    }

    sort.Slice(dirs, func(i, j int) bool {
        return dirs[i].Count > dirs[j].Count
    })

    // Return the directory with the highest count
    return dirs[0].Dir, nil
}

