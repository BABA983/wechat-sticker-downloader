package main

import (
	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const defaultTargetDirIndex = -1

var targetDirIndex = flag.Int("target-dir-index", defaultTargetDirIndex, "required to specify directory if multiple directories are found")

var outputDir string

func init() {
	const usage = "stickers output directory"
	defaultOutputDir := filepath.Join(os.Getenv("HOME"), "Downloads")
	flag.StringVar(&outputDir, "output-dir", defaultOutputDir, usage)
	flag.StringVar(&outputDir, "o", defaultOutputDir, usage+" (shorthand)")
}

func main() {
	flag.Parse()
	fmt.Println("targetDirIndex", *targetDirIndex)
	fmt.Println("outputDir", outputDir)
	// Find user data directory
	weChatDataPath := filepath.Join(os.Getenv("HOME"), "Library", "Containers", "com.tencent.xinWeChat", "Data", "Library", "Application Support", "com.tencent.xinWeChat", "2.0b4.0.9")
	dirs, _ := filepath.Glob(filepath.Join(weChatDataPath, "*"))
	dirs = mapper(dirs, func(s string) string { return filepath.Base(s) })
	dirs = filter(dirs, func(s string) bool { return len(s) > 30 })
	dirs = filter(dirs, func(dir string) bool {
		return fileExist(filepath.Join(weChatDataPath, dir, "Stickers", "fav.archive"))
	})
	dirsLength := len(dirs)
	if dirsLength == 0 {
		fmt.Println("No user directory found")
		return
	}

	// If there are multiple directories, ask user to specify
	if dirsLength > 1 && (*targetDirIndex < 1 || *targetDirIndex > len(dirs)) {
		fmt.Println("Multiple user directories found, please specify the index when running again:")
		for i, dir := range dirs {
			fmt.Printf("%d. %s\n", i+1, dir)
		}
		return
	}

	userDir := dirs[0]
	if dirsLength > 1 {
		userDir = dirs[*targetDirIndex-1]
	}
	// Copy fav.archive file to desktop
	sourcePath := filepath.Join(weChatDataPath, userDir, "Stickers", "fav.archive")
	destPath := filepath.Join(outputDir, "wechat-stickers.plist")
	copyFile(sourcePath, destPath)
	fmt.Println("fav.archive file copied successfully")
	fmt.Println("UserDir: ", userDir)
	fmt.Println("fav.archive path: ", sourcePath)
	fmt.Println("fav.archive dest path: ", destPath)

	// Convert fav.archive.plist to XML
	// cmd := "plutil -convert xml1 fav.archive.plist"
	cmd := "plutil -convert xml1 " + destPath
	execCmd(cmd)

	// Parse fav.archive.plist and extract image URLs
	input, _ := readFile(destPath)
	var urls []string
	decoder := xml.NewDecoder(bytes.NewReader(input))
	decoder.Strict = false

	for {
		token, err := decoder.Token()
		if token == nil {
			fmt.Println(err)
			break
		}
		switch start := token.(type) {
		case xml.StartElement:
			switch start.Name.Local {
			case "string":
				var s string
				err := decoder.DecodeElement(&s, &start)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				if strings.HasPrefix(s, "http") {
					urls = append(urls, s)
				}
			}
		}
	}

	fmt.Printf("%d emoji links found\n", len(urls))

	// Create download directory
	downloadDir := filepath.Join(outputDir, userDir+"-stickers")
	createDir(downloadDir)

	// Download emoji images
	fmt.Println("Starting to download emojis...")
	var wg sync.WaitGroup
	concurrentTasks := 50
	semaphore := make(chan struct{}, concurrentTasks)
	for _, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(url string) {
			defer wg.Done()
			name := generateRandomString(32)
			filename := filepath.Join(downloadDir, fmt.Sprintf("%s.gif", name))
			downloadFile(url, filename)
			fmt.Printf("Downloaded %s emoji\n", name)
			<-semaphore
		}(url)
	}

	wg.Wait()

	fmt.Println("All emojis downloaded to", outputDir)
}
