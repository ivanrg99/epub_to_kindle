package main

import (
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

type Kindle struct {
	Path string
}

func (k *Kindle) IsConnected() bool {
	partitions, _ := disk.Partitions(false)
	for _, partition := range partitions {
		path := fmt.Sprintf("%s/system/version.txt", partition.Mountpoint)
		f, err := os.Open(path)
		if err == nil {
			k.Path = partition.Mountpoint
			f.Close()
			return true
		}
	}
	k.Path = ""
	return false
}

func (k *Kindle) Process(path string) error {
	kindleGen := fmt.Sprintf("assets/%s/kindlegen", runtime.GOOS)
	pathBase := filepath.Base(path[:len(path)-5])
	res := fmt.Sprintf("%s.azw3", pathBase)
	//s, err := exec.Command(kindleGen, path, "-o", res).Output()
	err := exec.Command(kindleGen, path, "-o", res).Run()
	//fmt.Println(string(s))
	if err != nil {
		if err.Error() != "exit status 1" {
			fmt.Println("Kindle error")
			fmt.Println(err)
			return err
		}
	}

	fmt.Println(kindleGen)
	resFull := fmt.Sprintf("%s.azw3", path[:len(path)-5])
	dest := fmt.Sprintf("%s/documents/%s", k.Path, res)
	fmt.Println(resFull, dest)
	//err = MoveFile(resFull, dest)
	fmt.Println(err)
	fmt.Println(res)
	fmt.Println("DONE!")
	return nil
}

func MoveFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		in.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	in.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}

	err = out.Sync()
	if err != nil {
		return fmt.Errorf("Sync error: %s", err)
	}

	err = os.Remove(src)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}
