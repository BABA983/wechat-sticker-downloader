package main

import (
	crypto_rand "crypto/rand"
	"fmt"
	"io"
	math_rand "math/rand"
	"net/http"
	"os"
	"os/exec"
)

func readFile(path string) ([]byte, error) {
	file, err := os.Open(path) // For read access.
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}
	return data, nil
}

func fileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func mapper(s []string, f func(string) string) []string {
	var r []string
	for _, v := range s {
		r = append(r, f(v))
	}
	return r
}

func filter(s []string, f func(string) bool) []string {
	var r []string
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}

func copyFile(src, dst string) {
	out, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()
	in, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		fmt.Println("io.Copy err: ", err)
	}
}

func execCmd(cmd string) {
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("execCmd", string(output))
}

func createDir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.Mkdir(path, 0755)
	}
}

func downloadFile(url, filepath string) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.71 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("http.Client err: ", err)
		return
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		fmt.Println("os.Create err: ", err)
		return
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("io.Copy err: ", err)
	}
}

func generateRandomString(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"0123456789"
	b := make([]byte, n)
	crypto_rand.Read(b)
	for i := range b {
		b[i] = charset[math_rand.Intn(len(charset))]
	}
	return string(b)
}
