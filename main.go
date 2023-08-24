package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const (
	SPI_SETDESKWALLPAPER = 0x0014
	SPIF_UPDATEINIFILE   = 0x01
	SPIF_SENDCHANGE      = 0x02
)

func main() {
	//ticker := time.NewTicker(1 * time.Hour)
	//ticker := time.NewTicker(5 * time.Second)
	ticker := time.NewTicker(10 * time.Minute)

	baseDir := "./wallpapers"

	for range ticker.C {
		changeWallpaper(baseDir)
	}
}

func getTimeOfDay(t time.Time) string {
	hour := t.Hour()

	switch {
	case hour >= 5 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 17:
		return "noon"
	case hour >= 17 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func getRandomImageForTimeOfDay(baseDir, timeOfDay string) (string, error) {
	dirPath := filepath.Join(baseDir, timeOfDay)
	var images []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if the file has a .png extension
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".png") {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			images = append(images, absPath)
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no images found in %s", dirPath)
	}

	// return random image from the 'images' slice
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	return images[r.Intn(len(images))], nil

	// if you want to have static backgrounds, u can just return the first/only image in the slice
	// return images[0], nil
}

func setWallpaperStyleAndTile() error {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\Desktop`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	err = k.SetStringValue("WallpaperStyle", "6") // "6" is for the "Fit" style
	if err != nil {
		return err
	}

	err = k.SetStringValue("TileWallpaper", "0") // "0" means don't tile the wallpaper
	return err
}

func changeWallpaper(baseDir string) {
	currentTime := time.Now()

	timeOfDay := getTimeOfDay(currentTime)

	imagePath, err := getRandomImageForTimeOfDay(baseDir, timeOfDay)
	if err != nil {
		log.Fatalf("Failed to get random image: %v", err)
		return
	}

	err = setWallpaperStyleAndTile()
	if err != nil {
		log.Fatalf("Failed to set wallpaper style and tile: %v", err)
		return
	}

	// set the wallpaper
	ret, _, err := windows.NewLazySystemDLL("user32.dll").NewProc("SystemParametersInfoW").Call(
		SPI_SETDESKWALLPAPER,
		0,
		uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(imagePath))),
		SPIF_UPDATEINIFILE|SPIF_SENDCHANGE,
	)

	if ret == 0 {
		log.Fatalf("Failed to set wallpaper: %v", err)
	}
}
