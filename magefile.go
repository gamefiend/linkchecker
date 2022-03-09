//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/mg"
)

func Build() error {
	fmt.Println("Building...")
	cmd := exec.Command("go", "build", "-o", "./linkchecker", "./cmd/main.go")
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.Chmod("./linkchecker", 0755); err != nil {
		return err
	}

	return nil
}

func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing...")
	if err := os.Rename("./linkchecker", "/usr/bin/linkchecker"); err != nil {
		mg.Deps(Clean)
		return err
	}
	return nil
}

func Test() error {
	fmt.Println("Testing...")
	output, err := exec.Command("go", "test", "-v", "./...").Output()
	if err != nil {
		return err
	}
	fmt.Println(string(output))
	return nil
}
func Run() error {
	mg.Deps(Build)
	fmt.Println("Testing...")
	output, err := exec.Command("./linkchecker", "-v", "https://thoughtcrime-games.ghost.io").Output()
	if err != nil {
		mg.Deps(Clean)
		return err
	}
	mg.Deps(Clean)
	fmt.Println(string(output))
	return nil
}

func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("linkchecker")
}
