package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) CrackMnemonic(mnemonic, address string) string {
	result := BruteForceMnemonic(mnemonic, address)
	resultString := ""
	if result.Success {
		resultString += fmt.Sprintf("✅ 成功找到助记词: %s\n", result.Mnemonic)
		resultString += fmt.Sprintf("🔹 地址: %s\n", result.Address)
		resultString += fmt.Sprintf("⏱️ 耗时: %s\n", result.TimeElapsed)
		resultString += fmt.Sprintf("📊 处理数: %d/%d (%.2f%%)\n", result.Processed, result.Total, float64(result.Processed)/float64(result.Total)*100)
		resultString += fmt.Sprintf("⚡ 速率: %.2f 次/秒\n", result.Rate)
	} else {
		resultString += fmt.Sprintf("❌ 未找到匹配的助记词\n")
		resultString += fmt.Sprintf("⏱️ 耗时: %s\n", result.TimeElapsed)
		resultString += fmt.Sprintf("📊 处理数: %d/%d (%.2f%%)\n", result.Processed, result.Total, float64(result.Processed)/float64(result.Total)*100)
		resultString += fmt.Sprintf("⚡ 速率: %.2f 次/秒\n", result.Rate)
	}
	return resultString
}

func (a *App) StopBruteForce() {
	StopBruteForceMnemonic()
}

func (a *App) GetProgress() string {
	return PrintProgress()
}
