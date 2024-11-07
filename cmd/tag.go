package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	path string
	bkp  bool
)

func readFile() (*os.File, string, error) {
	var filePath string
	if path == "" {
		filePath = "./.env"
	} else {
		filePath = path
	}

	file, err := os.Open(filePath)

	if err != nil {
		return nil, "", fmt.Errorf("Failed to open the file: %v", err)
	}

	return file, filePath, nil
}

func replaceTag(modifiedLines []string, line string, key string, tag string) ([]string, error) {
	parts := strings.Split(line, "/")

	if len(parts) != 2 {
		return nil, errors.New("Unable to find the tag")
	}

	fileKey := strings.Split(parts[0], "=")[0]

	if strings.TrimSpace(fileKey) != key {
		return nil, errors.New("Unable to find the tag")
	}

	updated := parts[0] + "/" + tag
	modifiedLines = append(modifiedLines, updated)

	return modifiedLines, nil
}

func readAndReplace(file *os.File, key string, tag string) ([]string, error) {
	var modifiedLines []string
	var err error
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			modifiedLines = append(modifiedLines, line)
			continue
		}

		if strings.HasPrefix(line, key) {
			modifiedLines, err = replaceTag(modifiedLines, line, key, tag)

			if err != nil {
				fmt.Println(fmt.Errorf("%v", err))
				return nil, err
			}
		} else {
			modifiedLines = append(modifiedLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(fmt.Errorf("Failed to scan the file: %v", err))
		return nil, err
	}

	return modifiedLines, nil
}

func writeFile(filePath string, lines []string) {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(fmt.Errorf("Failed to open file: %v", err))
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			fmt.Println(fmt.Errorf("Failed to write to file: %v", err))
			return
		}
	}

	err = writer.Flush()
	if err != nil {
		fmt.Println(fmt.Errorf("Error flushing the file: %v", err))
		return
	}
}

func copyFile(file *os.File) (string, error) {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	bkpFileName := fmt.Sprintf("%s_bkp_%s", file.Name(), timestamp)

	destinationFile, err := os.Create(bkpFileName)

	if err != nil {
		return "", fmt.Errorf("Could not create destination file: %v", err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, file)

	if err != nil {
		return "", fmt.Errorf("Could not copy file content: %v", err)
	}

	_, err = file.Seek(0, io.SeekStart)

	if err != nil {
		return "", fmt.Errorf("Failed to reset the file after copy: %v", err)
	}

	return bkpFileName, nil
}

func removeFile(p string) error {
	err := os.Remove(p)

	if err != nil {
		return fmt.Errorf("Error deleting file: %v", err)
	}

	return nil
}

var tagCmd = &cobra.Command{
	Use:   "r",
	Short: "Replaces the tag in the env files",
	Long:  `This command allows quick switching of the environment tags in the remote environment for the purpose of deployment`,
	Run: func(cmd *cobra.Command, args []string) {
		key := args[1]
		tag := args[2]

		key = strings.TrimSpace(key)
		tag = strings.TrimSpace(tag)

		if len(args) != 3 {
			fmt.Printf("Invalid number of arguments. Command expects tag key as first argument and tag value as second argument. It also expects an optional third argument to the env file. Length: %v", args)
			return
		}

		file, filePath, err := readFile()

		if err != nil {
			fmt.Printf("%v", err)
			return
		}

		defer file.Close()

		var bkpPath string

		if bkp == true {
			bkpPath, err = copyFile(file)

			if err != nil {
				fmt.Printf("%v", err)
				return
			}
		}

		modifiedLines, err := readAndReplace(file, key, tag)

		if err != nil {
			fmt.Printf("%v", err)
			_ = removeFile(bkpPath)
			return
		}

		if len(modifiedLines) == 0 {
			fmt.Println("File not updated!")
			_ = removeFile(bkpPath)
			return
		}

		writeFile(filePath, modifiedLines)

		fmt.Println("File updated!")
	},
}

func Execute() {
	tagCmd.Flags().StringVarP(&path, "path", "p", "", "The path to the file")
	tagCmd.Flags().BoolVarP(&bkp, "bkp", "b", true, "Create a backup of the file before replacing")

	if err := tagCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
