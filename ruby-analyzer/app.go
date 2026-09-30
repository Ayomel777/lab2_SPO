package main

import (
	"context"
	"os"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"ruby-analyzer/internal/analyzer"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Analyze вызывается из JS: Analyze(source)
func (a *App) Analyze(source string) analyzer.Result {
	return analyzer.Analyze(source)
}

// SelectFile открывает нативный диалог выбора файла и возвращает путь.
// Если пользователь отменил выбор — вернёт пустую строку.
func (a *App) SelectFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Выберите Ruby-файл",
		Filters: []runtime.FileFilter{
			{DisplayName: "Ruby files", Pattern: "*.rb"},
		},
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}

// ReadFile нужен для кнопки «Открыть .rb»
func (a *App) ReadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
