package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/manifoldco/promptui"
	xdraw "golang.org/x/image/draw"
)

// Config holds compression settings
type Config struct {
	JpegQuality  int    `json:"jpeg_quality"`
	OutputFormat string `json:"output_format"` // "original", "jpeg", "png"
	MaxWidth     int    `json:"max_width"`     // 0 = disabled
}

// Defaults
var defaultConfig = Config{
	JpegQuality:  75,
	OutputFormat: "original",
	MaxWidth:     0,
}

var appConfig = defaultConfig

// Task represents an image to process
type Task struct {
	OriginalPath   string
	TempPath       string
	FileName       string
	OriginalSize   int64
	CompressedSize int64
	Savings        float64
	Status         string // "OK", "SKIP", "ERR"
}

func main() {
	// Carica config dalla directory utente standard
	loadConfig()

	// Se ci sono argomenti (drag & drop sull'icona), processali subito
	if len(os.Args) > 1 {
		runBatch(os.Args[1:])
		keepOpen()
		return
	}

	fmt.Println("==========================================")
	fmt.Println("   IMG COMPRESSOR - CLI v1.0")
	fmt.Println("==========================================")

	mainMenu()
}

// --- Menu System ---

func mainMenu() {
	for {
		prompt := promptui.Select{
			Label: "Seleziona un'azione",
			Items: []string{
				"Comprimi immagini (Drag & Drop)",
				"Configura parametri",
				"Exit",
			},
			Templates: getTemplates(),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return
		}

		switch result {
		case "Comprimi immagini (Drag & Drop)":
			compressionLoop()
		case "Configura parametri":
			configureParameters()
		case "Exit":
			os.Exit(0)
		}
	}
}

func configureParameters() {
	for {
		// Prepare labels for current values
		formatLabel := strings.ToUpper(appConfig.OutputFormat)
		resizeLabel := "Disabilitato"
		if appConfig.MaxWidth > 0 {
			resizeLabel = fmt.Sprintf("%d px", appConfig.MaxWidth)
		}

		items := []string{
			fmt.Sprintf("Qualità JPEG (Attuale: %d)", appConfig.JpegQuality),
			fmt.Sprintf("Formato Output (Attuale: %s)", formatLabel),
			fmt.Sprintf("Ridimensiona Max Width (Attuale: %s)", resizeLabel),
			"Ripristina Default",
			"Indietro",
		}

		prompt := promptui.Select{
			Label:     "Configurazione",
			Items:     items,
			Templates: getTemplates(),
		}

		i, _, err := prompt.Run()
		if err != nil {
			return
		}

		switch i {
		case 0: // JPEG Quality
			changeJpegQuality()
		case 1: // Output Format
			changeOutputFormat()
		case 2: // Max Width
			changeMaxWidth()
		case 3: // Reset
			appConfig = defaultConfig
			saveConfig()
			fmt.Println("✅ Configurazione ripristinata ai valori di default.")
		case 4: // Back
			return
		}
	}
}

func changeJpegQuality() {
	prompt := promptui.Prompt{
		Label:    "Nuova Qualità JPEG (1-100)",
		Default:  strconv.Itoa(appConfig.JpegQuality),
		Validate: validateNumber(1, 100),
	}
	res, _ := prompt.Run()
	val, _ := strconv.Atoi(res)
	appConfig.JpegQuality = val
	saveConfig()
}

func changeOutputFormat() {
	prompt := promptui.Select{
		Label:     "Seleziona Formato Output",
		Items:     []string{"Original (Mantieni formato)", "Force JPEG", "Force PNG"},
		Templates: getTemplates(),
	}
	i, _, _ := prompt.Run()
	switch i {
	case 0:
		appConfig.OutputFormat = "original"
	case 1:
		appConfig.OutputFormat = "jpeg"
	case 2:
		appConfig.OutputFormat = "png"
	}
	saveConfig()
}

func changeMaxWidth() {
	prompt := promptui.Prompt{
		Label:    "Larghezza Massima in Pixel (0 per disabilitare)",
		Default:  strconv.Itoa(appConfig.MaxWidth),
		Validate: validateNumber(0, 10000),
	}
	res, _ := prompt.Run()
	val, _ := strconv.Atoi(res)
	appConfig.MaxWidth = val
	saveConfig()
}

func getTemplates() *promptui.SelectTemplates {
	return &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "  {{ . | cyan }}", // Rimosso il carattere ▸ che a volte causa problemi
		Inactive: "  {{ . | faint }}",
		Selected: "✔ {{ . | white | bold }}",
	}
}

func validateNumber(min, max int) func(string) error {
	return func(input string) error {
		v, err := strconv.Atoi(input)
		if err != nil || v < min || v > max {
			return fmt.Errorf("inserisci un numero tra %d e %d", min, max)
		}
		return nil
	}
}

// --- Processing Logic ---

func compressionLoop() {
	fmt.Println("\n---------- Modalità Compressione ---------")
	fmt.Println("Trascina file o cartelle qui e premi INVIO.")
	fmt.Println("Scrivi 'back' per tornare al menu.")
	fmt.Println("------------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())

		if input == "back" || input == "exit" || input == "quit" {
			break
		}
		if input == "" {
			continue
		}

		paths := splitInput(input)
		runBatch(paths)

		// Post-compression menu
		prompt := promptui.Select{
			Label:     "Operazione completata. Cosa vuoi fare?",
			Items:     []string{"Carica altre immagini", "Torna al Menu Principale"},
			Templates: getTemplates(),
		}

		i, _, err := prompt.Run()
		if err != nil || i == 1 { // Error or "Torna al Menu Principale"
			break
		}
		// If i == 0 ("Carica altre immagini"), loop continues and shows "> " again
		fmt.Println("\nTrascina i prossimi file:")
	}
}

func runBatch(inputs []string) {
	fmt.Println("\n🔍 Analisi in corso...")

	var tasks []*Task
	totalOrig := int64(0)
	totalComp := int64(0)

	files := scanAll(inputs)
	if len(files) == 0 {
		fmt.Println("❌ Nessuna immagine supportata trovata.")
		return
	}

	for _, path := range files {
		task := analyzeImage(path)
		tasks = append(tasks, task)

		if task.Status == "OK" {
			totalOrig += task.OriginalSize
			totalComp += task.CompressedSize
		}
		fmt.Print(".")
	}
	fmt.Println("\n")

	printTable(tasks)

	savedBytes := totalOrig - totalComp
	savedPerc := 0.0
	if totalOrig > 0 {
		savedPerc = (float64(savedBytes) / float64(totalOrig)) * 100
	}

	fmt.Printf("\n📊 TOTALE ORIGINALE:  %s\n", formatBytes(totalOrig))
	fmt.Printf("📉 TOTALE STIMATO:    %s\n", formatBytes(totalComp))
	fmt.Printf("💰 RISPARMIO:         %s (%.1f%%)\n", formatBytes(savedBytes), savedPerc)

	if savedBytes <= 0 && appConfig.MaxWidth == 0 && appConfig.OutputFormat == "original" {
		fmt.Println("\nNessun risparmio significativo e nessuna conversione richiesta.")
		fmt.Println("Le immagini non verranno salvate.")
		cleanup(tasks)
		return
	}

	prompt := promptui.Prompt{
		Label:     "Procedere con il salvataggio",
		IsConfirm: true,
	}
	_, err := prompt.Run()

	if err != nil {
		fmt.Println("Operazione annullata.")
		cleanup(tasks)
		return
	}

	saveFiles(tasks)
}

func analyzeImage(path string) *Task {
	task := &Task{
		OriginalPath: path,
		FileName:     filepath.Base(path),
		Status:       "ERR",
	}

	info, err := os.Stat(path)
	if err != nil {
		return task
	}
	task.OriginalSize = info.Size()

	file, err := os.Open(path)
	if err != nil {
		return task
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return task
	}

	// 1. Resize Logic
	if appConfig.MaxWidth > 0 && img.Bounds().Dx() > appConfig.MaxWidth {
		ratio := float64(appConfig.MaxWidth) / float64(img.Bounds().Dx())
		newHeight := int(float64(img.Bounds().Dy()) * ratio)

		dst := image.NewRGBA(image.Rect(0, 0, appConfig.MaxWidth, newHeight))
		// Use CatmullRom for high quality resizing
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), xdraw.Over, nil)
		img = dst
	}

	// 2. Format Logic
	targetFormat := appConfig.OutputFormat
	if targetFormat == "original" {
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".png" {
			targetFormat = "png"
		} else {
			targetFormat = "jpeg"
		}
	}

	// Update filename extension if format changed
	if targetFormat == "jpeg" && !strings.HasSuffix(strings.ToLower(task.FileName), ".jpg") && !strings.HasSuffix(strings.ToLower(task.FileName), ".jpeg") {
		task.FileName = strings.TrimSuffix(task.FileName, filepath.Ext(task.FileName)) + ".jpg"
	} else if targetFormat == "png" && !strings.HasSuffix(strings.ToLower(task.FileName), ".png") {
		task.FileName = strings.TrimSuffix(task.FileName, filepath.Ext(task.FileName)) + ".png"
	}

	// Temp File
	tempFile, err := os.CreateTemp("", "imgcomp-*.tmp")
	if err != nil {
		return task
	}
	task.TempPath = tempFile.Name()
	defer tempFile.Close()

	// 3. Compression Logic
	switch targetFormat {
	case "jpeg":
		// Handle transparency for PNG -> JPEG conversion
		if _, ok := img.(*image.RGBA); ok {
			// Create white background
			newImg := image.NewRGBA(img.Bounds())
			draw.Draw(newImg, newImg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
			draw.Draw(newImg, newImg.Bounds(), img, image.Point{}, draw.Over)
			img = newImg
		}
		err = jpeg.Encode(tempFile, img, &jpeg.Options{Quality: appConfig.JpegQuality})
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		err = encoder.Encode(tempFile, img)
	}

	if err != nil {
		return task
	}

	stat, _ := tempFile.Stat()
	task.CompressedSize = stat.Size()

	// Logic: If user specifically requested resize or format change, we save even if size increases.
	// Otherwise (pure compression), we skip if size increases.
	forcedChange := (appConfig.MaxWidth > 0) || (appConfig.OutputFormat != "original")

	if !forcedChange && task.CompressedSize >= task.OriginalSize {
		task.Status = "SKIP"
		task.CompressedSize = task.OriginalSize
	} else {
		task.Status = "OK"
		task.Savings = 100 - (float64(task.CompressedSize)/float64(task.OriginalSize))*100
	}

	return task
}

func scanAll(inputs []string) []string {
	var files []string
	for _, input := range inputs {
		pathStr := cleanPath(input)
		info, err := os.Stat(pathStr)
		if err != nil {
			fmt.Println("⚠️ Errore lettura:", pathStr)
			continue
		}
		if info.IsDir() {
			filepath.Walk(pathStr, func(p string, i os.FileInfo, err error) error {
				if err == nil && !i.IsDir() && isImage(p) {
					files = append(files, p)
				}
				return nil
			})
		} else if isImage(pathStr) {
			files = append(files, pathStr)
		}
	}
	return files
}

func isImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png"
}

func saveFiles(tasks []*Task) {
	count := 0
	for _, task := range tasks {
		if task.Status != "OK" {
			os.Remove(task.TempPath)
			continue
		}
		srcDir := filepath.Dir(task.OriginalPath)
		outDir := filepath.Join(srcDir, "compressed")
		os.MkdirAll(outDir, 0755)

		destPath := filepath.Join(outDir, task.FileName)
		err := os.Rename(task.TempPath, destPath)
		if err != nil {
			copyFile(task.TempPath, destPath)
			os.Remove(task.TempPath)
		}
		count++
	}
	fmt.Printf("\n✅ Completato! %d immagini salvate.\n", count)
}

func cleanup(tasks []*Task) {
	for _, task := range tasks {
		if task.TempPath != "" {
			os.Remove(task.TempPath)
		}
	}
}

func printTable(tasks []*Task) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "FILE\tORIGINALE\tNUOVO\tRISPARMIO\tSTATO")
	fmt.Fprintln(w, "----\t---------\t-----\t---------\t-----")
	for _, t := range tasks {
		orig := formatBytes(t.OriginalSize)
		comp := formatBytes(t.CompressedSize)
		perc := fmt.Sprintf("-%.1f%%", t.Savings)
		if t.Status == "SKIP" {
			comp = "-"
			perc = "0%"
		}
		name := t.FileName
		if len(name) > 20 {
			name = name[:17] + "..."
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, orig, comp, perc, t.Status)
	}
	w.Flush()
}

// --- Config Storage ---

func getConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "config.json" // Fallback locale
	}
	appDir := filepath.Join(configDir, "ImgCompress")
	os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "config.json")
}

func loadConfig() {
	path := getConfigPath()
	file, err := os.Open(path)
	if err != nil {
		return // Usa default
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&appConfig)
}

func saveConfig() {
	path := getConfigPath()
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Errore salvataggio config:", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(appConfig)
	fmt.Println("Configurazione salvata in:", path)
}

// --- Utils ---

func splitInput(input string) []string {
	var args []string
	var current strings.Builder
	escaped := false
	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		if r == ' ' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteRune(r)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

func cleanPath(path string) string {
	path = strings.Trim(path, "\"'")
	return strings.TrimSpace(path)
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}

func keepOpen() {
	fmt.Println("\nPremi Invio per chiudere...")
	fmt.Scanln()
}
