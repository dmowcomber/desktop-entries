package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/alediaferia/prefixmap"
	"gopkg.in/ini.v1"
)

type DesktopEntry struct {
	Config struct {
		Name string
		Exec string
	}
	Path string
}

func main() {
	prefixMap := getDesktopEntries()

	// take user input
	fmt.Print("Search an Application (Desktop Entry): ")
	reader := bufio.NewReader(os.Stdin)
	searchString, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("input error: %s", err.Error())
	}
	searchString = strings.ToLower(searchString)

	printResults(searchString, prefixMap)
}

func printResults(searchString string, prefixMap *prefixmap.PrefixMap) {
	fmt.Printf("\n\nResults matching %s\n", searchString)
	results := prefixMap.GetByPrefix(searchString)
	for _, res := range results {
		de := res.(DesktopEntry)
		fmt.Printf("App Name: %s\n", de.Config.Name)
		fmt.Printf("\t.desktop file path: %s\n", de.Path)
		fmt.Printf("\t.exec: %s\n", de.Config.Exec)
	}
}

func getDesktopEntries() *prefixmap.PrefixMap {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("warning: unable to get the user's home directory: %s", err.Error())
	}

	dirs := []string{
		homeDir + "/.local/share/applications",
		homeDir + "/Desktop",
		"/usr/share/applications",
		"/usr/local/share/applications/",
	}

	prefixMap := prefixmap.New()
	for _, dir := range dirs {
		err := filepath.WalkDir(dir,
			func(path string, dirEntry fs.DirEntry, err error) error {
				if err != nil {
					fmt.Printf("walk error: %s - %T - %s\n", path, err, err.Error())
					if errors.Is(err, os.ErrNotExist) {
						return nil
					}
				}
				if !strings.HasSuffix(strings.ToLower(dirEntry.Name()), ".desktop") {
					return nil
				}
				if err != nil {
					return err
				}
				cfg, err := ini.Load(path)
				if err != nil {
					fmt.Printf("Fail to read file: %v", err)
					os.Exit(1)
				}
				for _, section := range cfg.Sections() {
					if section.Name() != "Desktop Entry" {
						continue
					}

					// fmt.Println(section.KeyStrings())
					de := DesktopEntry{
						Path: path,
					}
					de.Config.Name = section.Key("Name").String()
					de.Config.Exec = section.Key("Exec").String()

					prefixMap.Insert(strings.ToLower(de.Config.Name), de) // TODO: do I need to lowercase the key?
				}
				return nil
			})
		if err != nil {
			log.Println(err)
		}
	}
	return prefixMap
}
