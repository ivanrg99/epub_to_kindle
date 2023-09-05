package main

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
	"path/filepath"
	"time"
)

type App struct {
	Font     map[int]rl.Font
	Width    int32
	Height   int32
	Kindle   Kindle
	Progress TransferProgress
}

type TransferProgress struct {
	InProgress      bool
	Complete        chan bool
	ItemsInProgress int
	FailedFiles     []string
	CompletedFiles  []string
}

// NewFontMap MUST be called after rl.InitWindow
func NewFontMap(font string) map[int]rl.Font {
	const chars = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHI\nJKLMNOPQRSTUVWXYZ[]^_`abcdefghijklmn\nopqrstuvwxyz{|}~¿ÀÁÂÃÄÅÆÇÈÉÊËÌÍÎÏÐÑÒÓ\nÔÕÖ×ØÙÚÛÜÝÞßàáâãäåæçèéêëìíîïðñòóôõö÷\nøùúûüýþÿ"
	sizes := []int{10, 16, 20, 24, 32, 48, 62, 72}
	fonts := make(map[int]rl.Font, len(sizes))
	for _, i := range sizes {
		fonts[i] = rl.LoadFontEx(font, int32(i), []rune(chars))
	}
	return fonts
}

func Init() *App {
	const screenWidth = 800
	const screenHeight = 450

	rl.InitWindow(screenWidth, screenHeight, "Kindle Converter")
	rl.SetTargetFPS(60)

	font := NewFontMap("assets/font.ttf")

	progress := TransferProgress{
		InProgress:      false,
		Complete:        make(chan bool),
		ItemsInProgress: 0,
		FailedFiles:     make([]string, 0),
		CompletedFiles:  make([]string, 0),
	}

	app := App{
		Font:   font,
		Width:  screenWidth,
		Height: screenHeight,
		Kindle: Kindle{
			Path: "",
		},
		Progress: progress,
	}
	return &app
}

func DrawCenterText(app *App, text string, size float32, color rl.Color) {
	r := rl.MeasureTextEx(app.Font[int(size)], text, size, 0)
	pos := rl.Vector2{
		X: float32(app.Width/2) - (r.X / 2),
		Y: float32(app.Height/2) - r.Y,
	}
	rl.DrawTextEx(app.Font[int(size)], text, pos, size, 0, color)
}

func UpdateLoop(app *App) {
	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		if !app.Kindle.IsConnected() {
			DrawCenterText(app, "Connect your Kindle!", 32, rl.Black)
		} else {
			//TODO: Extract logic into separate functions
			if app.Progress.InProgress {
				select {
				case <-app.Progress.Complete:
					app.Progress.ItemsInProgress--
					if app.Progress.ItemsInProgress == 0 {
						app.Progress.InProgress = false
						if len(app.Progress.FailedFiles) > 0 {
							pos := rl.Vector2{
								X: 100,
								Y: 40,
							}
							rl.DrawTextEx(app.Font[24], "Failed to convert these files:", pos, 24, 0, rl.Black)

							y := 0
							for i, f := range app.Progress.FailedFiles {
								if i%2 == 0 {
									rl.DrawRectangle(0, int32(85+40*i), app.Width, 40, rl.Fade(rl.LightGray, 0.5))
								} else {
									rl.DrawRectangle(0, int32(85+40*i), app.Width, 40, rl.Fade(rl.LightGray, 0.3))
								}
								y = 100 + 40*i
								pos := rl.Vector2{
									X: 120,
									Y: float32(y),
								}
								rl.DrawTextEx(app.Font[20], f, pos, 20, 0, rl.Red)
							}

							pos = rl.Vector2{
								X: 100,
								Y: float32(y + 80),
							}
							if len(app.Progress.CompletedFiles) > 0 {
								rl.DrawTextEx(app.Font[24], "The following files were sent:", pos, 24, 0, rl.Black)

								for i, f := range app.Progress.CompletedFiles {
									if i%2 == 0 {
										rl.DrawRectangle(0, int32(y+85+40*(i+1)), app.Width, 40, rl.Fade(rl.LightGray, 0.5))
									} else {
										rl.DrawRectangle(0, int32(y+85+40*(i+1)), app.Width, 40, rl.Fade(rl.LightGray, 0.3))
									}
									pos := rl.Vector2{
										X: 120,
										Y: float32(140 + y + 40*i),
									}
									rl.DrawTextEx(app.Font[20], f, pos, 20, 0, rl.DarkGreen)
								}

							}

						} else {
							DrawCenterText(app, "Done!", 32, rl.DarkGreen)
						}
						rl.EndDrawing()
						time.Sleep(5 * time.Second)
						app.Progress.FailedFiles = make([]string, 0)
						app.Progress.CompletedFiles = make([]string, 0)
					}
				default:
					DrawCenterText(app, "Converting...", 32, rl.Black)
				}
			} else {
				DrawCenterText(app, "Drop EPUBs!", 32, rl.Black)
				if rl.IsFileDropped() {
					droppedFiles := rl.LoadDroppedFiles()
					app.Progress.ItemsInProgress = len(droppedFiles)
					for _, f := range droppedFiles {
						fmt.Println("Dropped: ", f)
						go func(f string) {
							err := app.Kindle.Process(f)
							if err != nil {
								fmt.Println("Found error")
								app.Progress.FailedFiles = append(app.Progress.FailedFiles, filepath.Base(f))
							} else {
								app.Progress.CompletedFiles = append(app.Progress.CompletedFiles, filepath.Base(f))
							}
							app.Progress.Complete <- true
						}(f)
					}
					app.Progress.InProgress = true
				}
			}
		}
		rl.EndDrawing()
	}
	rl.CloseWindow()
}
