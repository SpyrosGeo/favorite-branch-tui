package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)
var app *tview.Application
var statusBar *tview.TextView
func main() {
	// Initialize TUI application
	app = tview.NewApplication()

	// Create a list to display branches
	branchesList := tview.NewList().
		AddItem("Loading branches...", "", ' ', nil)

	// Create a status bar for displaying messages
	statusBar = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("")

	// Fetch and display Git branches
	updateBranchesList(branchesList)

	// Set up event handler for selecting branches
	branchesList.SetSelectedFunc(func(index int, label string, secondaryText string, shortcut rune) {
		branchName := strings.TrimSpace(label)
		// Checkout to the selected branch
		checkoutBranch(branchName)
		// Save the selected branch to favorites
		saveFavoriteBranch(branchName)
	})

	// Set up keybinding to save current branch and quit app
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 's':
				// Get the current branch
				currentBranch, err := getCurrentBranch()
				if err != nil {
					showMessage("Error getting current branch: " + err.Error())
					return event
				}
				if currentBranch != "" {
					// Save the current branch to favorites
					saveFavoriteBranch(currentBranch)
					// Update the branches list
					updateBranchesList(branchesList)
				}
			case 'q':
				// Quit the application
				app.Stop()
			}
		}
		return event
	})

	// Create the layout
	flex := tview.NewFlex().
		AddItem(branchesList, 0, 1, true).
		AddItem(statusBar, 1, 0, false)

	// Set the root layout and run the application
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

func updateBranchesList(branchesList *tview.List) {
	// Open favorite branches file
	file, err := os.OpenFile("favorite_branches.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		showMessage("Error opening file: " + err.Error())
		return
	}
	defer file.Close()

	// Maintain a set to store unique branch names
	seen := make(map[string]bool)

	// Read branch names from file
	scanner := bufio.NewScanner(file)
	empty := true
	for scanner.Scan() {
		branchName := strings.TrimSpace(scanner.Text())
		if branchName != "" && !seen[branchName] {
			branchesList.AddItem(branchName, "", 0, nil)
			seen[branchName] = true
			empty = false
		}
	}

	if err := scanner.Err(); err != nil {
		showMessage("Error reading file: " + err.Error())
		return
	}

	if empty {
		branchesList.AddItem("No favorite branches", "", 0, nil)
	}
}

func checkoutBranch(branchName string) {
	// Execute "git checkout" command to switch branches
	cmd := exec.Command("git", "checkout", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		showMessage("Error: " + err.Error())
		return
	}
	showMessage(string(output))
}

func saveFavoriteBranch(branchName string) {
	// Open file in append mode
	file, err := os.OpenFile("favorite_branches.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		showMessage("Error opening file: " + err.Error())
		return
	}
	defer file.Close()

	// Write branch name to file if it doesn't already exist
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		seen[scanner.Text()] = true
	}

	if !seen[branchName] {
		writer := bufio.NewWriter(file)
		_, err = writer.WriteString(branchName + "\n")
		if err != nil {
			showMessage("Error writing to file: " + err.Error())
			return
		}

		// Flush the writer
		err = writer.Flush()
		if err != nil {
			showMessage("Error flushing writer: " + err.Error())
			return
		}
	}
}

func getCurrentBranch() (string, error) {
	// Execute "git rev-parse --abbrev-ref HEAD" command to get the current branch
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func showMessage(message string) {
	app.QueueUpdateDraw(func() {
		statusBar.SetText(message)
	})
}
