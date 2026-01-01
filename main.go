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
	"runtime"
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
	// Load config from standard user config directory
	loadConfig()

	// If arguments are present (drag & drop onto icon), process them immediately
	if len(os.Args) > 1 {
		runBatch(os.Args[1:])
		keepOpen()
		return
	}

	fmt.Println("================================================")
	fmt.Println("   IMG COMPRESSOR - CLI v1.0.1 by V. Antedoro")
	fmt.Println("   https://github.com/antedoro/imgcompress-go")
	fmt.Println("================================================")

	mainMenu()
}

// --- Menu System ---

func mainMenu() {
	for {
		prompt := promptui.Select{
			Label: "Select an action",
			Items: []string{
				"Compress images (Drag & Drop)",
				"Configure parameters",
				"Exit",
			},
			Templates: getTemplates(),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return
		}

		switch result {
		case "Compress images (Drag & Drop)":
			compressionLoop()
		case "Configure parameters":
			configureParameters()
		case "Exit":
			fmt.Println("✔ Exit")
			fmt.Println("Press enter to exit.")
			bufio.NewReader(os.Stdin).ReadString('\n')
			os.Exit(0)
		}
	}
}

func configureParameters() {
	for {
		// Prepare labels for current values
		formatLabel := strings.ToUpper(appConfig.OutputFormat)
		resizeLabel := "Disabled"
		if appConfig.MaxWidth > 0 {
			resizeLabel = fmt.Sprintf("%d px", appConfig.MaxWidth)
		}

		items := []string{
			fmt.Sprintf("JPEG Quality (Current: %d)", appConfig.JpegQuality),
			fmt.Sprintf("Output Format (Current: %s)", formatLabel),
			fmt.Sprintf("Resize Max Width (Current: %s)", resizeLabel),
			"Restore Defaults",
			"Back",
		}

		prompt := promptui.Select{
			Label:     "Configuration",
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
			fmt.Println("✅ Configuration restored to defaults.")
		case 4: // Back
			return
		}
	}
}

func changeJpegQuality() {
	prompt := promptui.Prompt{
		Label:    "New JPEG Quality (1-100)",
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
		Label:     "Select Output Format",
		Items:     []string{"Original (Keep format)", "Force JPEG", "Force PNG"},
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
		Label:    "Maximum Width in Pixels (0 to disable)",
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
		Label:    "{{ . | cyan }}",
		Active:   "  {{ . | cyan }} ",
		Inactive: "  {{ . | faint }} ",
		Selected: "✔ {{ . | white | bold }} ",
	}
}

func validateNumber(min, max int) func(string) error {
	return func(input string) error {
		v, err := strconv.Atoi(input)
		if err != nil || v < min || v > max {
			return fmt.Errorf("please enter a number between %d and %d", min, max)
		}
		return nil
	}
}

// --- Processing Logic ---

func compressionLoop() {
	fmt.Println("\n---------- Compression Mode --------------")
	fmt.Println("Drag files or folders here and press ENTER.")
	fmt.Println("Type 'back' to return to the menu.")
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
			Label:     "Operation completed. What would you like to do?",
			Items:     []string{"Load more images", "Return to Main Menu"},
			Templates: getTemplates(),
		}

		i, _, err := prompt.Run()
		if err != nil || i == 1 { // Error or "Return to Main Menu"
			break
		}
		// If i == 0 ("Load more images"), loop continues and shows "> " again
		fmt.Println("\nDrag the next files:")
	}
}

func runBatch(inputs []string) {
	fmt.Println("\n🔍 Analyzing...")

	var tasks []*Task
	totalOrig := int64(0)
	totalComp := int64(0)

	files := scanAll(inputs)
	if len(files) == 0 {
		fmt.Println("❌ No supported images found.")
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

	fmt.Printf("\n📊 TOTAL ORIGINAL:  %s\n", formatBytes(totalOrig))
	fmt.Printf("📉 TOTAL ESTIMATED:  %s\n", formatBytes(totalComp))
	fmt.Printf("💰 SAVINGS:          %s (%.1f%%)\n", formatBytes(savedBytes), savedPerc)

	if savedBytes <= 0 && appConfig.MaxWidth == 0 && appConfig.OutputFormat == "original" {
		fmt.Println("\nNo significant savings and no conversion requested.")
		fmt.Println("Images will not be saved.")
		cleanup(tasks)
		return
	}

	prompt := promptui.Prompt{
		Label:     "Proceed with saving",
		IsConfirm: true,
	}
	_, err := prompt.Run()

	if err != nil {
		fmt.Println("Operation cancelled.")
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
			fmt.Println("⚠️  Read error:", pathStr)
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
	fmt.Printf("\n✅ Completed! %d images saved.\n", count)
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
	fmt.Fprintln(w, "FILE\tORIGINAL\tNEW\tSAVINGS\tSTATUS")
	fmt.Fprintln(w, "----\t--------\t---\t-------\t------")
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
		return "config.json" // Fallback local
	}
	appDir := filepath.Join(configDir, "ImgCompress")
	os.MkdirAll(appDir, 0755)
	return filepath.Join(appDir, "config.json")
}

func loadConfig() {
	path := getConfigPath()
	file, err := os.Open(path)
	if err != nil {
		return // Use defaults
	}
	defer file.Close()
	json.NewDecoder(file).Decode(&appConfig)
}

func saveConfig() {
	path := getConfigPath()
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error saving config:", err)
		return
	}
	defer file.Close()
	json.NewEncoder(file).Encode(appConfig)
	fmt.Println("Configuration saved to:", path)
}

// --- Utils ---

func splitInput(input string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	escaped := false

	for _, r := range input {
		if escaped {
			current.WriteRune(r)
			escaped = false
			continue
		}

		if runtime.GOOS != "windows" && r == '\\' {
			escaped = true
			continue
		}

		if r == '"' {
			inQuotes = !inQuotes
			continue
		}

		if r == ' ' && !inQuotes {
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
	fmt.Println("\nPress Enter to close...")
	fmt.Scanln()
}
