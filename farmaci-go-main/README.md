# Farmaci-GO v_0.1

by Vincenzo Antedoro

Applicazione per generare un messaggio di richiesta farmaci.

## Come usarla

Esegui l'applicazione con:

```bash
go run main.go
```

Usa i tasti freccia e Invio per navigare il menu e selezionare pazienti e farmaci.

## Build

Per compilare l'applicazione per diverse piattaforme, usa i seguenti comandi:

**Windows (64-bit):**
```bash
GOOS=windows GOARCH=amd64 go build -o farmaci-go.exe
```

**macOS (Intel):**
```bash
GOOS=darwin GOARCH=amd64 go build -o farmaci-go-intel
```

**macOS (Apple Silicon):**
```bash
GOOS=darwin GOARCH=arm64 go build -o farmaci-go-arm64
```

**Linux (64-bit):**
```bash
GOOS=linux GOARCH=amd64 go build -o farmaci-go
```