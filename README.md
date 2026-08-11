# UK-US TTS Player

App simples para uso pessoal feito em Go com (Fyne). Este app recebe um texto em inglês e reproduz com sotaque britânico e americano usando **Piper TTS** + **mpv**.

## Pré-requisitos

1. **Piper** instalado — baixe o binário para Linux em:
   https://github.com/rhasspy/piper/releases
   Extraia de forma que o binário fique acessível em `./piper/piper` (relativo à pasta do projeto), ou ajuste a constante `piperBin` em `main.go`.

2. **Vozes britânicas** (`.onnx` + `.onnx.json`) — baixe de:
   https://github.com/rhasspy/piper/blob/master/VOICES.md
   Procure por vozes com prefixo `en_GB-` (ex: `en_GB-alba-medium`, `en_GB-alan-medium`).
   Coloque os arquivos em `./piper-voices/`.

   Se os nomes dos arquivos que você baixar forem diferentes dos listados no mapa `britishVoices` em `main.go`, ajuste esse mapa com os caminhos corretos.

3. **mpv** instalado e no `PATH`:
   ```bash
   # Void Linux
   sudo xbps-install -S mpv

   # Debian/Ubuntu
   sudo apt install mpv
   ```

4. **Dependências do Fyne** (bibliotecas gráficas do sistema):
   ```bash
   # Void Linux
   sudo xbps-install -S gcc pkg-config libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-devel libXi-devel libXxf86vm-devel
   ```

## Estrutura esperada

```
desktop-audio-english-US-GB/
├── main.go
├── go.mod
├── piper/
│   └── piper              (binário)
├── piper-voices/
│   ├── en_GB-alba-medium.onnx
│   ├── en_GB-alba-medium.onnx.json
│   └── ...
└── audios/                 (gerado automaticamente)
```

## Rodando

```bash
go mod tidy
go run main.go
```

## Uso

1. Digite o texto em inglês no campo.
2. Escolha a voz britânica desejada.
3. Clique em **Play** — o app gera o áudio com o Piper e reproduz automaticamente via mpv.

Os arquivos gerados ficam salvos em `./audios/`, nomeados a partir do texto (minúsculo, sem espaços/acentos), para evitar duplicação de trabalho em textos repetidos.

## Notas

- O Piper sempre gera arquivos **WAV**, mesmo que você peça outro nome — por isso o app já força a extensão `.wav`.
- Se o botão Play não reagir ou o áudio não tocar, rode o app pelo terminal (`go run main.go`) para ver mensagens de erro do Piper ou do mpv diretamente no log.
- Para trocar de voz, adicione entradas no mapa `britishVoices` em `main.go` apontando para o caminho do `.onnx` correspondente.

## install
```bash
curl -fsSL https://raw.githubusercontent.com/DavidSilva-S/desktop-audio-english-US-GB/main/install.sh | sh
```

