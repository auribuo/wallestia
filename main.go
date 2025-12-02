package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	ps "github.com/mitchellh/go-ps"
)

func getWallPid() (int, error) {
	processList, err := ps.Processes()
	if err != nil {
		return 0, err
	}

	for _, proc := range processList {
		if proc.Executable() == "linux-wallpaperengine" {
			return proc.Pid(), nil
		}
	}
	return -1, nil
}

func setWall(wp string) error {
	pid, err := getWallPid()
	if err != nil {
		return err
	}
	if pid >= 0 {
		unsetWall()
	}
	proc := exec.Command("hyprctl",
		"hyprctl",
		"dispatch",
		"--",
		"exec",
		"linux-wallpaperengine",
		"--fps",
		"30",
		"--scaling",
		"fill",
		"--screen-root",
		"eDP-1",
		wp,
	)

	_ = proc.Wait()
	return nil
}

func unsetWall() error {
	pid, err := getWallPid()
	if err != nil {
		return err
	}
	if pid < 0 {
		return fmt.Errorf("Process linux-wallpaperengine not running")
	}

	proc, _ := os.FindProcess(pid)
	return proc.Kill()
}

func onWallUpdate(home string) {
	caelestia_path := filepath.FromSlash(home)
	caelestia_path = filepath.Join(caelestia_path, ".local", "state", "caelestia", "wallpaper", "path.txt")
	wall_path, err := os.ReadFile(caelestia_path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Failed to read file %s: %s\n", caelestia_path, err.Error())
		os.Exit(1)
	}
	wall_filename := filepath.Base(string(wall_path))
	wall_id := strings.Split(wall_filename, ".")[0]

	steam_path := filepath.FromSlash(home)
	steam_path = filepath.Join(".steam", "steam", "steamapps", "Workshop", "Content", "431960")

	wall_file_maybe := filepath.Join(steam_path, wall_id)
	stat, err := os.Stat(wall_file_maybe)
	if err != nil || !stat.IsDir() {
		fmt.Fprintf(os.Stderr, "warning: Wallpaper %s does not exist\n", wall_file_maybe)
		err = unsetWall()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Failed to unset wallpaper: %s\n", err.Error())
		}
	} else {
		err = setWall(wall_file_maybe)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: Failed to set wallpaper: %s\n", err.Error())
		}
	}
}

func main() {
	home := os.Getenv("HOME")
	if home == "" {
		fmt.Fprintf(os.Stderr, "error: $HOME env variable not set!")
		os.Exit(1)
	}
	caelestia_path := filepath.FromSlash(home)
	caelestia_path = filepath.Join(caelestia_path, ".local", "state", "caelestia", "wallpaper", "path.txt")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Failed to create fsnotify watcher: %s\n", err.Error())
		os.Exit(1)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				fmt.Println("Got event")
				if !ok {
					fmt.Println("Got error")
					return
				}

				if event.Has(fsnotify.Write) {
					onWallUpdate(home)
				}
			}
		}
	}()

	err = watcher.Add(caelestia_path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: Failed to create fsnotify watch: %s\n", err.Error())
		os.Exit(1)
	}

	for {
	}
}
