# ImgCompress-go

**ImgCompress-go** è un'applicazione desktop leggera e potente per ottimizzare, ridimensionare e convertire immagini (JPEG, PNG). Scritta in Go, offre un'interfaccia a riga di comando (CLI) interattiva e supporta il Drag & Drop nativo.

![Icon](icon.png)

## ✨ Funzionalità

*   **Compressione Intelligente:** Riduce il peso delle immagini mantenendo alta la qualità visiva.
*   **Ridimensionamento:** Imposta una larghezza massima (es. 1920px) e l'app ridimensionerà le immagini più grandi proporzionalmente.
*   **Conversione:** Converti facilmente da PNG a JPG (o viceversa) o mantieni il formato originale.
*   **Drag & Drop:** Trascina file o intere cartelle sull'icona dell'app o dentro la finestra.
*   **Analisi Preventiva:** Vedi quanto spazio risparmierai *prima* di confermare il salvataggio.
*   **Configurazione Persistente:** Le tue preferenze vengono salvate automaticamente.

## 🚀 Installazione e Uso

### macOS
1.  Scarica il file `ImgCompress.app`.
2.  Trascinalo nella cartella **Applicazioni** (o tienilo sul Desktop).
3.  **Uso:**
    *   **Drag & Drop:** Trascina le immagini direttamente sull'icona dell'app.
    *   **Doppio Click:** Apri l'app per configurare i parametri (Qualità, Formato, Resize) tramite il menu interattivo.

### Windows
1.  Scarica `ImgCompress_Win_x64.exe`.
2.  Esegui il file dal Prompt dei Comandi o tramite doppio click.

## 🛠 Compilazione (Build)

Se vuoi compilare il progetto dal codice sorgente:

### Prerequisiti
*   [Go](https://go.dev/dl/) (v1.21+) installato.

### Comandi

1.  Clona il repository:
    ```bash
    git clone https://github.com/tuo-username/imgcompress-go.git
    cd imgcompress-go
    ```

2.  Installa le dipendenze:
    ```bash
    go mod tidy
    ```

3.  **Compilazione per macOS:**
    ```bash
    chmod +x build_mac.sh
    ./build_mac.sh
    ```
    Troverai `ImgCompress.app` nella cartella corrente.

4.  **Compilazione per Windows:**
    ```bash
    chmod +x build_win.sh
    ./build_win.sh
    ```
    Troverai gli eseguibili nella cartella `dist/`.

## ⚙️ Configurazione

Le impostazioni vengono salvate in:
*   **macOS:** `~/Library/Application Support/ImgCompress/config.json`
*   **Windows:** `%AppData%\ImgCompress\config.json`

Esempio di config:
```json
{
  "jpeg_quality": 75,
  "output_format": "original",
  "max_width": 1920
}
```

## 📝 Licenza

Distribuito sotto licenza MIT. Vedi `LICENSE` per maggiori informazioni.
