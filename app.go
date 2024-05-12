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

	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}
	statusBar = tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("")

		// Fetch and display Git branches
		// Create the layout
	branchesList, err := getSavedBranches()
	if err != nil {
		panic(err)
	}

	pages := tview.NewPages()
	pages.AddPage("background", branchesList, true, true).AddPage("modal", modal(statusBar, 40, 10), true, true)

	// Set the root layout and run the application
	if err := app.SetRoot(branchesList, true).Run(); err != nil {
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
	fmt.Println("inside checkoutBranch")
	showMessage(branchName)
	// cmd := exec.Command("git", "checkout", branchName)
	// output, err := cmd.CombinedOutput()
	// if err != nil {
	// 	showMessage("Error: " + err.Error())
	// 	return
	// }
	// showMessage(string(output))
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
	// file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	return err
	// }
	// defer file.Close()
	//
	// // Write the text to the file
	// _, err = file.WriteString(currentBranch)
	// if err != nil {
	// 	return err
	// }
	//
	// return nil
	// Read the current contents of the file
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

	// Add a "Quit" item
	list.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})
	list.AddItem("Save", "Save branch to Favorites", 's', func() {
		saveBranchToFavorites()
	})

	return list, nil
}
