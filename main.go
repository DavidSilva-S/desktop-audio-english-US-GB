package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)



const appName = "desktop-audio-english-US-GB"

const mpvBin = "mpv" 


type appDirs struct {
	piperBin  string // binário do piper
	voicesDir string // pasta com os .onnx + .onnx.json
	audiosDir string // pasta de saída dos áudios gerados (cache)
}

func resolveAppDirs() (appDirs, error) {
	dataHome, err := os.UserHomeDir()
	if err != nil {
		return appDirs{}, fmt.Errorf("não foi possível resolver o diretório home: %w", err)
	}
	dataDir := filepath.Join(dataHome, ".local", "share", appName)

	cacheHome, err := os.UserCacheDir()
	if err != nil {
		return appDirs{}, fmt.Errorf("não foi possível resolver o diretório de cache: %w", err)
	}
	audiosDir := filepath.Join(cacheHome, appName, "audios")

	dirs := appDirs{
		piperBin:  filepath.Join(dataDir, "piper", "piper"),
		voicesDir: filepath.Join(dataDir, "piper-voices"),
		audiosDir: audiosDir,
	}

		if err := os.MkdirAll(dirs.audiosDir, 0755); err != nil {
		return appDirs{}, fmt.Errorf("erro ao criar pasta de áudios: %w", err)
	}

	return dirs, nil
}

func checkInstallation(dirs appDirs) error {
	if _, err := os.Stat(dirs.piperBin); os.IsNotExist(err) {
		return fmt.Errorf(
			"piper não encontrado em %s\nRode o instalador novamente: curl -fsSL <URL_DO_INSTALL_SH> | sh",
			dirs.piperBin,
		)
	}
	entries, err := os.ReadDir(dirs.voicesDir)
	if err != nil || len(entries) == 0 {
		return fmt.Errorf(
			"nenhuma voz encontrada em %s\nRode o instalador novamente: curl -fsSL <URL_DO_INSTALL_SH> | sh",
			dirs.voicesDir,
		)
	}
	return nil
}

var englishVoices = map[string]string{
	"Alba (feminina) - British":            "./piper-voices/en_GB-alba-medium.onnx",
	"Alan (masculina) - British":           "./piper-voices/en_GB-alan-medium.onnx",
	"Cori (feminina) - British":            "./piper-voices/en_GB-cori-medium.onnx",
	"Librits (feminina) - American":        "./piper-voices/en_US-libritts-high.onnx",	
}


func generateAudio(dirs appDirs, text, voiceFile, fileName string) (string, error) {
	modelPath := filepath.Join(dirs.voicesDir, voiceFile)
	outputPath := filepath.Join(dirs.audiosDir, fileName)

	cmd := exec.Command(dirs.piperBin, "--model", modelPath, "--output_file", outputPath)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("erro ao abrir stdin do piper: %w", err)
	}

	var stderrBuf strings.Builder
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("erro ao iniciar piper (verifique o caminho do binário): %w", err)
	}

	if _, err := io.WriteString(stdin, text); err != nil {
		stdin.Close()
		return "", fmt.Errorf("erro ao escrever texto no piper: %w", err)
	}
	stdin.Close()

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("piper falhou: %w\n%s", err, stderrBuf.String())
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("piper não gerou o arquivo esperado em %s", outputPath)
	}

	return outputPath, nil
}

func playAudio(path string) error {
	cmd := exec.Command(mpvBin, "--no-terminal", path)
	return cmd.Start()
}

func safeFileName(text string) string {
	clean := strings.ToLower(text)
	clean = strings.ReplaceAll(clean, " ", "_")
	
	var b strings.Builder
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	name := b.String()
	if name == "" {
		name = "audio"
	}
	if len(name) > 40 {
		name = name[:40]
	}
	return name + ".wav"
}

func main() {
	a := app.NewWithID("com.david.uktts")
	win := a.NewWindow("UK-US TTS Player")
	win.Resize(fyne.NewSize(480, 260))

	title := widget.NewLabel("Text to US-UK voice ")
	title.Alignment = fyne.TextAlignCenter

	statusLabel := widget.NewLabel("")
	statusLabel.Alignment = fyne.TextAlignCenter

	dirs, err := resolveAppDirs()
	if err != nil {
		statusLabel.SetText("Erro fatal: " + err.Error())
	} else if err := checkInstallation(dirs); err != nil {
		statusLabel.SetText(err.Error())
	}

	textEntry := widget.NewMultiLineEntry()
	textEntry.SetPlaceHolder("Digite o texto em inglês aqui...")
	textEntry.Wrapping = fyne.TextWrapWord
	textEntry.SetMinRowsVisible(4)

	voiceNames := make([]string, 0, len(englishVoices))
	for name := range englishVoices {
		voiceNames = append(voiceNames, name)
	}
	voiceSelect := widget.NewSelect(voiceNames, func(string) {})
	if len(voiceNames) > 0 {
		voiceSelect.SetSelected(voiceNames[0])
	}

	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	var btnPlay *widget.Button
	btnPlay = widget.NewButtonWithIcon("Play", theme.MediaPlayIcon(), func() {
		text := strings.TrimSpace(textEntry.Text)
		if text == "" {
			statusLabel.SetText("Digite um texto antes de tocar.")
			return
		}

		voiceFile, ok := englishVoices[voiceSelect.Selected]
		if !ok {
			statusLabel.SetText("Selecione uma voz válida.")
			return
		}

		if err := checkInstallation(dirs); err != nil {
			statusLabel.SetText(err.Error())
			return
		}

		btnPlay.Disable()
		progress.Show()
		statusLabel.SetText("Gerando áudio...")

		go func() {
			defer func() {
				progress.Hide()
				btnPlay.Enable()
			}()

			fileName := safeFileName(text)
			path, err := generateAudio(dirs, text, voiceFile, fileName)
			if err != nil {
				statusLabel.SetText("Erro ao gerar áudio: " + err.Error())
				return
			}

			statusLabel.SetText("Reproduzindo...")
			if err := playAudio(path); err != nil {
				statusLabel.SetText("Erro ao reproduzir: " + err.Error())
				return
			}
			statusLabel.SetText("Pronto.")
		}()
	})
	btnPlay.Importance = widget.HighImportance

	controls := container.NewVBox(
		title,
		textEntry,
		container.NewBorder(nil, nil, widget.NewLabel("Voz:"), nil, voiceSelect),
		container.NewCenter(btnPlay),
		progress,
		statusLabel,
	)

	win.SetContent(controls)
	win.ShowAndRun()
}
