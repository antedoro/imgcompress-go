This is a CLI application written in Go. It helps users generate a message to request prescription drugs for multiple patients.

The application has the following features:
- A list of patients with their pre-defined drug lists.
- A menu to select patients and their specific drugs.
- A feature to generate a formatted message with the selected drugs for the selected patients.
- The ability to copy the generated message to the clipboard.

The `go.mod` file shows dependencies on `github.com/atotto/clipboard` for clipboard operations and `github.com/manifoldco/promptui` for the interactive CLI prompts.

The `main.go` file contains the core logic:
- It defines `Drug` and `Patient` structs.
- It initializes a list of patients with their drugs.
- The `main` function is a loop that presents a menu with options: "Select Patients/Drugs", "Generate Message", and "Exit".
- `selectPatientsAndDrugs` function allows the user to select a patient and then select the drugs for that patient.
- `selectDrugs` function allows the user to select drugs for a patient.
- `generateMessage` function constructs the message with the selected patients and drugs and provides an option to copy it to the clipboard.
