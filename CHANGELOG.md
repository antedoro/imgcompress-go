# Changelog

Tutte le modifiche notevoli a questo progetto saranno documentate in questo file.

## [1.0.0] - 2026-01-01

Prima versione stabile rilasciata pubblicamente.

### 🚀 Nuove Funzionalità
- **Interfaccia Interattiva:** Menu CLI navigabile con tastiera per configurazione e avvio processi.
- **Supporto Drag & Drop:**
  - Trascina file/cartelle sull'icona dell'app (macOS) per processare immediatamente.
  - Trascina file nella finestra del terminale aperta.
- **Motore di Compressione:**
  - Supporto ottimizzato per **JPEG** (qualità configurabile).
  - Supporto **PNG** (massima compressione).
  - Gestione trasparenza automatica (sfondo bianco) nelle conversioni PNG -> JPEG.
- **Ridimensionamento (Resize):**
  - Opzione `Max Width` per ridimensionare automaticamente immagini troppo grandi mantenendo l'aspect ratio (algoritmo Catmull-Rom di alta qualità).
- **Conversione Formati:**
  - Possibilità di forzare l'output in JPEG o PNG, oppure mantenere il formato originale.
- **Analisi e Anteprima:**
  - Tabella dettagliata pre-salvataggio che mostra dimensione originale, stimata e percentuale di risparmio.
  - Calcolo totale dei MB risparmiati.
  - Skip automatico se il file compresso risulta più grande dell'originale.
- **Configurazione Persistente:**
  - Salvataggio preferenze utente (Qualità, Formato, Resize) in `config.json` nelle cartelle di sistema (`Application Support` / `AppData`).
- **Build System:**
  - Script `build_mac.sh` per generare Bundle `.app` nativi con icona personalizzata e launcher terminale.
  - Script `build_win.sh` per cross-compilare eseguibili Windows (x64 e ARM64).

### 🛠 Correzioni e Miglioramenti
- Risolto problema di chiusura immediata del terminale su macOS.
- Gestione corretta dei percorsi con spazi e caratteri speciali (escape).
- Eliminazione dei segnali acustici (beep) durante la navigazione nei menu.
