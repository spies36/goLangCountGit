package main

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"
)

type childChannel struct {
	repoCount   (chan int)
	foundString (chan string)
}

func main() {

	var cwd string
	var err error
	var folders []fs.DirEntry
	var parentChannel = childChannel{
		repoCount:   make(chan int, 1),
		foundString: make(chan string, 1),
	}

	cwd, err = os.Getwd()

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	fmt.Printf("Beginning search from path: %v \n", cwd)
	startTime := time.Now()

	folders, err = findAndCountRootDirs()

	if err != nil {
		fmt.Print(err.Error())
		return
	}

	go countGitRepos(folders, cwd, parentChannel)

	fmt.Printf("%v", <-parentChannel.foundString)
	fmt.Printf("Found %d git repos in %f seconds \n", <-parentChannel.repoCount, time.Since(startTime).Seconds())
}

// Count all of the directories and return them
func findAndCountRootDirs() (folders []fs.DirEntry, err error) {
	folders, err = os.ReadDir("./")
	if err != nil {
		return
	}
	return
}

func countGitRepos(directoryEntries []fs.DirEntry, directoryPath string, channel childChannel) {
	var repoCount int = 0
	var foundString string = ""
	var childChannels = []childChannel{}

	for _, dirEntry := range directoryEntries {

		if !dirEntry.IsDir() { // ignore
			continue
		} else if isGitRepo(dirEntry.Name()) { //found repo
			repoCount++
			foundString += fmt.Sprintf("Git Repo: %v \n", directoryPath)
		} else { //explore directory
			var subDirectoryPath = path.Join(directoryPath, dirEntry.Name())
			var subDirEntries, err = os.ReadDir(subDirectoryPath)
			if err != nil {
				fmt.Printf("Error reading directory %v \n", subDirectoryPath)
				continue
			}
			newChan := childChannel{
				repoCount:   make(chan int),
				foundString: make(chan string),
			}

			childChannels = append(childChannels, newChan)

			go countGitRepos(subDirEntries, subDirectoryPath, newChan)
		}
	}

	for _, childChannel := range childChannels {
		repoCount += <-childChannel.repoCount
		foundString += <-childChannel.foundString
	}
	channel.repoCount <- repoCount
	channel.foundString <- foundString
	defer close(channel.repoCount)
	defer close(channel.foundString)
}

func isGitRepo(folderName string) bool {
	return strings.ToLower(folderName) == ".git"
}
