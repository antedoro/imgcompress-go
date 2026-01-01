package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/manifoldco/promptui"
)

type Drug struct {
	Name     string `json:"Name"`
	Selected bool   `json:"Selected"`
}

type Patient struct {
	Name     string `json:"Name"`
	Drugs    []Drug `json:"Drugs"`
	Selected bool   `json:"Selected"`
}

func main() {
	fmt.Println("Farmaci-GO v_0.1 by Vincenzo Antedoro")
	fmt.Println("Applicazione per generare un messaggio di richiesta farmaci.")
	fmt.Println()

	patients := loadPatients()

	for {
		prompt := promptui.Select{
			Label:     "Select an action",
			Items:     []string{"Select Patients/Drugs", "Generate Message", "Exit"},
			Templates: getSelectTemplates(),
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			fmt.Println("Press enter to exit.")
			fmt.Scanln()
			return
		}

		switch result {
		case "Select Patients/Drugs":
			patients = selectPatientsAndDrugs(patients)
			savePatients(patients)
		case "Generate Message":
			generateMessage(patients)
		case "Exit":
			for i := range patients {
				patients[i].Selected = false
				for j := range patients[i].Drugs {
					patients[i].Drugs[j].Selected = false
				}
			}
			savePatients(patients)
			fmt.Println("Press enter to exit.")
			fmt.Scanln()
			os.Exit(0)
		}
	}
}

func getSelectTemplates() *promptui.SelectTemplates {
	return &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "▸ {{ . | cyan }}",
		Inactive: "  {{ . | faint }}",
		Selected: "✔ {{ . | white | bold }}",
	}
}

func getDataPath() string {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return "data.json"
	}
	exeDir := filepath.Dir(exePath)

	// Check if running via 'go run'
	if strings.Contains(exeDir, "go-build") {
		return "data.json" // In 'go run' case, data.json is in the current working directory
	}

	// when running as .app, the executable is in Contents/MacOS, and data is in Contents/Resources
	if strings.Contains(exeDir, ".app/Contents/MacOS") {
		return filepath.Join(exeDir, "..", "Resources", "data.json")
	}

	// For standalone executable, assume data.json is in the same directory as the executable
	return filepath.Join(exeDir, "data.json")
}

func loadPatients() []Patient {
	dataPath := getDataPath()
	data, err := ioutil.ReadFile(dataPath)
	if err != nil {
		fmt.Printf("Error reading data.json: %v\n", err)
		return []Patient{}
	}

	var patients []Patient
	err = json.Unmarshal(data, &patients)
	if err != nil {
		fmt.Printf("Error unmarshalling data.json: %v\n", err)
		return []Patient{}
	}

	return patients
}

func savePatients(patients []Patient) {
	dataPath := getDataPath()
	data, err := json.MarshalIndent(patients, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling patients: %v\n", err)
		return
	}

	err = ioutil.WriteFile(dataPath, data, 0644)
	if err != nil {
		fmt.Printf("Error writing to data.json: %v\n", err)
	}
}

func selectPatientsAndDrugs(patients []Patient) []Patient {
	patientNames := make([]string, len(patients))
	for i, p := range patients {
		status := " "
		if p.Selected {
			status = "✔"
		}
		patientNames[i] = fmt.Sprintf("[%s] %s", status, p.Name)
	}

	prompt := promptui.Select{
		Label:     "Select a patient to toggle selection or view drugs",
		Items:     patientNames,
		Templates: getSelectTemplates(),
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return patients
	}

	patients[i].Selected = !patients[i].Selected
	if patients[i].Selected {
		patients[i].Drugs = selectDrugs(patients[i].Drugs)
		// if no drugs are selected, deselect the patient
		hasSelectedDrugs := false
		for _, drug := range patients[i].Drugs {
			if drug.Selected {
				hasSelectedDrugs = true
				break
			}
		}
		if !hasSelectedDrugs {
			patients[i].Selected = false
		}
	} else {
		// Deselect all drugs if patient is deselected
		for j := range patients[i].Drugs {
			patients[i].Drugs[j].Selected = false
		}
	}

	return patients
}

func selectDrugs(drugs []Drug) []Drug {
	cursorPos := 0
	for {
		drugNames := make([]string, len(drugs)+2)
		for i, d := range drugs {
			status := " "
			if d.Selected {
				status = "✔"
			}
			drugNames[i] = fmt.Sprintf("[%s] %s", status, d.Name)
		}
		drugNames[len(drugs)] = "Manage Drugs"
		drugNames[len(drugs)+1] = "Done"

		prompt := promptui.Select{
			Label:     "Select a drug to toggle selection",
			Items:     drugNames,
			Size:      len(drugNames),
			CursorPos: cursorPos,
			Templates: getSelectTemplates(),
		}

		i, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return drugs
		}

		cursorPos = i

		switch result {
		case "Manage Drugs":
			drugs = manageDrugs(drugs)
		case "Done":
			return drugs
		default:
			drugs[i].Selected = !drugs[i].Selected
		}
	}
}

func manageDrugs(drugs []Drug) []Drug {
	for {
		prompt := promptui.Select{
			Label:     "Manage Drugs",
			Items:     []string{"Add Drug", "Change Drug", "Remove Drug", "Back"},
			Templates: getSelectTemplates(),
		}

		_, result, err := prompt.Run()

		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return drugs
		}

		switch result {
		case "Add Drug":
			drugs = addDrug(drugs)
		case "Change Drug":
			drugs = changeDrug(drugs)
		case "Remove Drug":
			drugs = removeDrug(drugs)
		case "Back":
			return drugs
		}
	}
}

func addDrug(drugs []Drug) []Drug {
	prompt := promptui.Prompt{
		Label: "Drug Name",
	}

	name, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return drugs
	}

	return append(drugs, Drug{Name: name})
}

func changeDrug(drugs []Drug) []Drug {
	drugNames := make([]string, len(drugs))
	for i, d := range drugs {
		drugNames[i] = d.Name
	}

	promptSelect := promptui.Select{
		Label:     "Select a drug to change",
		Items:     drugNames,
		Templates: getSelectTemplates(),
	}

	i, _, err := promptSelect.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return drugs
	}

	promptPrompt := promptui.Prompt{
		Label:   "New Drug Name",
		Default: drugs[i].Name,
	}

	name, err := promptPrompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return drugs
	}

	drugs[i].Name = name
	return drugs
}

func removeDrug(drugs []Drug) []Drug {
	drugNames := make([]string, len(drugs))
	for i, d := range drugs {
		drugNames[i] = d.Name
	}

	prompt := promptui.Select{
		Label:     "Select a drug to remove",
		Items:     drugNames,
		Size:      len(drugNames),
		Templates: getSelectTemplates(),
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return drugs
	}

	return append(drugs[:i], drugs[i+1:]...)
}

func generateMessage(patients []Patient) {
	var text strings.Builder
	text.WriteString("Gentile segreteria (dott. Pittari),\n")
	text.WriteString("chiedo la prescrizione dei seguenti farmaci\n")
	text.WriteString("da inviare a antedoro@gmail.com\n\n")

	hasSelectedDrugs := false
	for _, patient := range patients {
		if patient.Selected {
			selectedDrugs := []Drug{}
			for _, drug := range patient.Drugs {
				if drug.Selected {
					selectedDrugs = append(selectedDrugs, drug)
				}
			}

			if len(selectedDrugs) > 0 {
				hasSelectedDrugs = true
				text.WriteString(fmt.Sprintf("*%s:*\n", patient.Name))
				for _, drug := range selectedDrugs {
					text.WriteString(fmt.Sprintf("- %s\n", drug.Name))
				}
				text.WriteString("\n")
			}
		}
	}

	if !hasSelectedDrugs {
		fmt.Println("Nessun farmaco selezionato.")
		return
	}

	text.WriteString("Grazie\nVincenzo Antedoro")

	message := text.String()
	fmt.Println("-- Generated Message --")
	fmt.Println(message)
	fmt.Println("-------------------------")

	prompt := promptui.Select{
		Label:     "What would you like to do with the message?",
		Items:     []string{"Copy to Clipboard", "Done"},
		Templates: getSelectTemplates(),
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if result == "Copy to Clipboard" {
		clipboard.WriteAll(message)
		fmt.Println("Message copied to clipboard!")
	}
}