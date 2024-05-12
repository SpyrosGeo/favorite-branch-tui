package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/rivo/tview"
)

var app *tview.Application
var statusBar *tview.TextView

func main() {
	// Initialize TUI application
	app = tview.NewApplication()

	//Fetch and display Git branches
	branchesList, err := getSavedBranches()
	//Fetch general commands
	generalCommandsList, cmdErr := getGeneralCommands()
	if err != nil || cmdErr != nil {
		panic(err)
	}
	//createn primitive to use for secitons
	newPrimitive := func(text string) tview.Primitive {
		return tview.NewTextView().
			SetTextAlign(tview.AlignCenter).
			SetText(text)
	}

	title := getTitle()
	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0, 30).
		SetBorders(true).
		AddItem(newPrimitive(title), 0, 0, 1, 3, 0, 0, false).
		AddItem(generalCommandsList, 2, 0, 1, 3, 0, 0, false)
	grid.AddItem(branchesList, 1, 0, 1, 1, 0, 100, false)

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}

func checkoutBranch(branchName string) {
	// Execute "git checkout" command to switch branches
	fmt.Println("inside checkoutBranch")
	cmd := exec.Command("git", "checkout", branchName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		showMessage("Error: " + err.Error())
		return
	}
	showMessage(string(output))
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

func saveBranchToFavorites() {
	currentBranch, err := getCurrentBranch()
	if err != nil {
		panic(err)
	}
	writeErr := writeToDB(currentBranch)
	if writeErr != nil {
		panic(writeErr)
	}

}

func writeToDB(currentBranch string) error {
	filename := "favorite_branches.txt"
	currentContent, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Append the new text below the current content
	newContent := append(currentContent, []byte("\n"+currentBranch)...)

	// Write the updated content back to the file
	err = os.WriteFile(filename, newContent, 0644)
	if err != nil {
		return err
	}

	return nil
}

func getSavedBranches() (*tview.List, error) {
	// Open favorite branches file
	file, err := os.Open("favorite_branches.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Create a new list
	list := tview.NewList()

	// Maintain a set of encountered branch names to ignore duplicates
	encountered := make(map[string]bool)

	// Read branch names from file
	scanner := bufio.NewScanner(file)
	index := 'a'
	for scanner.Scan() {
		branchName := strings.TrimSpace(scanner.Text())
		if branchName != "" && !encountered[branchName] {
			// Add branch name to set
			encountered[branchName] = true

			// Create a closure for the callback to capture the branchName
			func(branchName string) {
				list.AddItem(branchName, "", index, func() {
					checkoutBranch(branchName)
				})
			}(branchName)
			index++
		}
	}

	return list, nil
}
func getGeneralCommands() (*tview.List, error) {

	list := tview.NewList()
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})
	list.AddItem("Save", "Save branch to Favorites", 's', func() {
		saveBranchToFavorites()
	})

	return list, nil
}
func getTitle() string {
	currentBranch, err := getCurrentBranch()
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("Current branch: %s", currentBranch)
}
