//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/magefile/mage/mg"
)

func Build() error {
	fmt.Println(color.GreenString("Building..."))
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
	fmt.Println(color.BlackString("Testing..."))
	cmd := exec.Command("sh", "-c", "go test -json |gotestdox")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()

}

func Run() error {
	mg.Deps(Build)
	fmt.Println(color.GreenString("Running..."))
	cmd := exec.Command("./linkchecker", "-v", "https://thoughtcrime-games.ghost.io")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	mg.Deps(Clean)
	return err
}

func RunJSON() error {
	mg.Deps(Build)
	fmt.Println(color.GreenString("Running with JSON Output..."))
	output, err := exec.Command("./linkchecker", "-v", "-j", "https://thoughtcrime-games.ghost.io").Output()
	if err != nil {
		mg.Deps(Clean)
		return err
	}
	fmt.Println(string(output))
	mg.Deps(Clean)
	return nil
}

func Tidy() error {
	fmt.Println(color.YellowString("Tidying..."))
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
func Clean() {
	fmt.Println("Cleaning...")
	os.RemoveAll("linkchecker")
}
