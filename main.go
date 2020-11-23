package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/joho/godotenv"
)

func main() {
	errLog := log.New(os.Stderr, "ERROR:\t", log.Lshortfile)

	err := godotenv.Load()
	if err != nil {
		errLog.Fatalln(err)
	}

	wpDir, ok := os.LookupEnv("WP_PATH")
	if !ok {
		errLog.Fatalln("Missing WP_PATH environment variable.")
	}

	files, err := ioutil.ReadDir(wpDir)
	if err != nil {
		errLog.Fatalln(err)
	}

	var wallpapers []string
	for _, file := range files {
		ext := filepath.Ext(file.Name())
		if ext == ".jpg" || ext == ".png" {
			wallpapers = append(wallpapers, file.Name())
		}
	}

	cmd := exec.Command(
		`osascript`,
		`-e`,
		`tell app "finder" to get posix path of (get desktop picture as alias)`,
	)

	var outb, errb bytes.Buffer
	cmd.Stdout, cmd.Stderr = &outb, &errb

	if err := cmd.Run(); err != nil {
		log.Fatalln(errb.String())
	}

	sliceOutb := strings.Split(outb.String(), "/")
	currentWallpaper := sliceOutb[len(sliceOutb)-1]

	activeRow := func() int {
		for i, wallpaper := range wallpapers {
			if strings.Contains(currentWallpaper, wallpaper) {
				return i
			}
		}
		return 0
	}()

	if err := ui.Init(); err != nil {
		errLog.Fatalln(err)
	}
	defer ui.Close()

	list := widgets.NewList()
	list.Title = "Wallpapers"
	list.Rows = wallpapers
	list.SelectedRow = activeRow
	list.TextStyle = ui.NewStyle(ui.ColorYellow)
	list.WrapText = false
	list.BorderStyle.Fg = ui.ColorYellow
	list.SetRect(0, 0, 25, 25)

	uiEvents := ui.PollEvents()
	ui.Render(list)

	for {
		e := <-uiEvents
		switch e.ID {
		case "<C-c>", "q":
			return
		case "j", "Up":
			list.ScrollDown()
		case "k", "Down":
			list.ScrollUp()
		default:
			continue
		}

		newWallpaper := filepath.Join(wpDir, wallpapers[list.SelectedRow])

		osaCmd := fmt.Sprintf(`tell application "Finder" to set desktop picture to POSIX file "%s"`, newWallpaper)
		err := exec.Command(`osascript`, `-e`, osaCmd).Run()
		if err != nil {
			errLog.Fatalln(err)
		}
		ui.Render(list)
	}
}
