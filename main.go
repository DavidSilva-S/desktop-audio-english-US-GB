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

// ---------------------------------------------------------------
// CONFIGURAÇÃO — ajuste os caminhos conforme sua instalação local
// ---------------------------------------------------------------
const (
	piperBin    = "./piper/piper"                              // binário do piper
	audiosDir   = "./audios"                                    // pasta de saída dos áudios
	mpvBin      = "mpv"                                         // player (precisa estar no PATH)
)

// Vozes britânicas conhecidas do Piper (baixe o .onnx + .onnx.json correspondente
// e coloque em ./piper-voices/). Ajuste os nomes conforme os arquivos que você tiver.
var britishVoices = map[string]string{
	"Alba (feminina) - British":            "./piper-voices/en_GB-alba-medium.onnx",
	"Alan (masculina) - British":           "./piper-voices/en_GB-alan-medium.onnx",
	"Cori (feminina) - British":            "./piper-voices/en_GB-cori-medium.onnx",
	"Librits (feminina) - American":        "./piper-voices/en_US-libritts-high.onnx",	
}

// generateAudio roda o Piper com o modelo escolhido e salva o WAV resultante.
// Piper sempre gera WAV — por isso o arquivo de saída usa extensão .wav.
func generateAudio(text, modelPath, fileName string) (string, error) {
	if err := os.MkdirAll(audiosDir, 0755); err != nil {
		return "", fmt.Errorf("erro ao criar pasta de áudios: %w", err)
	}

	outputPath := filepath.Join(audiosDir, fileName)

	cmd := exec.Command(piperBin, "--model", modelPath, "--output_file", outputPath)

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

// playAudio toca o WAV usando mpv em background.
func playAudio(path string) error {
	cmd := exec.Command(mpvBin, "--no-terminal", path)
	return cmd.Start()
}

// safeFileName normaliza o texto para um nome de arquivo seguro e consistente
// (tudo minúsculo, sem espaços) — evita o clássico bug de "Fruit.wav" vs "fruit.wav".
func safeFileName(text string) string {
	clean := strings.ToLower(text)
	clean = strings.ReplaceAll(clean, " ", "_")
	// remove qualquer coisa que não seja letra/número/underscore
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
	win := a.NewWindow("UK TTS Player")
	win.Resize(fyne.NewSize(480, 260))

	title := widget.NewLabel("Texto em inglês → voz britânica")
	title.Alignment = fyne.TextAlignCenter

	textEntry := widget.NewMultiLineEntry()
	textEntry.SetPlaceHolder("Digite o texto em inglês aqui...")
	textEntry.Wrapping = fyne.TextWrapWord
	textEntry.SetMinRowsVisible(4)

	// monta as opções do seletor de voz a partir do mapa
	voiceNames := make([]string, 0, len(britishVoices))
	for name := range britishVoices {
		voiceNames = append(voiceNames, name)
	}
	voiceSelect := widget.NewSelect(voiceNames, func(string) {})
	if len(voiceNames) > 0 {
		voiceSelect.SetSelected(voiceNames[0])
	}

	status := widget.NewLabel("")
	status.Alignment = fyne.TextAlignCenter

	progress := widget.NewProgressBarInfinite()
	progress.Hide()

	var btnPlay *widget.Button
	btnPlay = widget.NewButtonWithIcon("Play", theme.MediaPlayIcon(), func() {
		text := strings.TrimSpace(textEntry.Text)
		if text == "" {
			status.SetText("Digite um texto antes de tocar.")
			return
		}

		modelPath, ok := britishVoices[voiceSelect.Selected]
		if !ok {
			status.SetText("Selecione uma voz válida.")
			return
		}

		btnPlay.Disable()
		progress.Show()
		status.SetText("Gerando áudio...")

		go func() {
			defer func() {
				progress.Hide()
				btnPlay.Enable()
			}()

			fileName := safeFileName(text)
			path, err := generateAudio(text, modelPath, fileName)
			if err != nil {
				status.SetText("Erro ao gerar áudio: " + err.Error())
				return
			}

			status.SetText("Reproduzindo...")
			if err := playAudio(path); err != nil {
				status.SetText("Erro ao reproduzir: " + err.Error())
				return
			}
			status.SetText("Pronto.")
		}()
	})
	btnPlay.Importance = widget.HighImportance

	controls := container.NewVBox(
		title,
		textEntry,
		container.NewBorder(nil, nil, widget.NewLabel("Voz:"), nil, voiceSelect),
		container.NewCenter(btnPlay),
		progress,
		status,
	)

	win.SetContent(controls)
	win.ShowAndRun()
}
