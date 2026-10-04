package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func ClearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	} else {
		fmt.Print("\033[H\033[2J")
	}
}

func TuiPause(msg ...string) {
	prompt := "Tekan Enter untuk melanjutkan..."
	if len(msg) > 0 && msg[0] != "" {
		prompt = msg[0]
	}
	fmt.Println()
	fmt.Print(TextDim(fmt.Sprintf("  [ %s ] ", prompt)))
	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}

func TuiInput(prompt string, defVal ...string) string {
	defaultStr := ""
	if len(defVal) > 0 {
		defaultStr = defVal[0]
	}

	hint := ""
	if defaultStr != "" {
		hint = fmt.Sprintf(" [%s]", defaultStr)
	}

	fmt.Printf("  %s%s %s ", Primary("▸"), TextWhite(" "+prompt+hint), TextDim("❯"))
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultStr
	}
	res := strings.TrimSpace(line)
	if res == "" {
		return defaultStr
	}
	return res
}

func TuiConfirm(question string, defVal bool) bool {
	hint := "[y/N]"
	if defVal {
		hint = "[Y/n]"
	}

	fmt.Printf("  %s%s %s %s ", Amber("?"), TextWhite(" "+question), TextDim(hint), TextDim("❯"))
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return defVal
	}

	ans := strings.ToLower(strings.TrimSpace(line))
	if ans == "" {
		return defVal
	}
	return ans == "y" || ans == "ya" || ans == "yes"
}

func GetPromptText() string {
	return Primary("engsel") + " " + TextDim("❯") + " "
}
