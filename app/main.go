package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// aliases stores alias-to-command mappings
var aliases = make(map[string]string)

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		// Get current working directory for prompt
		wd, err := os.Getwd()
		if err != nil {
			wd = "/unknown"
		}
		// Get current user to determine home directory
		currentUser, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error getting current user:", err)
			continue
		}
		// Replace /home/user with ~
		homeDir := currentUser.HomeDir
		displayWd := wd
		if strings.HasPrefix(wd, homeDir) {
			displayWd = "~" + strings.TrimPrefix(wd, homeDir)
		}
		fmt.Fprintf(os.Stdout, "\033[36m%s\n\033[35m❯\033[0m ", displayWd)

		// Read user input
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}
		if err := execInput(command); err != nil {
			fmt.Fprintln(os.Stderr, "Command error:", err)
		}
	}
}

var ErrNoPath = errors.New("path required")

func execInput(input string) error {
	// Remove the newline character
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}

	// Split the input to separate the command and arguments
	args := strings.Fields(input)
	if len(args) == 0 {
		return nil
	}

	// Handle alias command (e.g., alias ll='ls -l')
	if args[0] == "alias" {
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: alias name='command'")
			return nil
		}
		// Split the alias definition
		aliasParts := strings.SplitN(args[1], "=", 2)
		if len(aliasParts) < 2 {
			fmt.Fprintln(os.Stderr, "Invalid alias format. Use: alias name='command'")
			return nil
		}
		name := strings.TrimSpace(aliasParts[0])
		cmd := strings.Trim(strings.TrimSpace(aliasParts[1]), "'\"")
		if name == "" || cmd == "" {
			fmt.Fprintln(os.Stderr, "Alias name or command cannot be empty")
			return nil
		}
		aliases[name] = cmd
		fmt.Fprintf(os.Stdout, "Alias '%s' set to '%s'\n", name, cmd)
		return nil
	}

	// Check if the command is an alias
	if aliasedCmd, exists := aliases[args[0]]; exists {
		// Replace the alias with the actual command
		if len(args) > 1 {
			input = aliasedCmd + " " + strings.Join(args[1:], " ")
		} else {
			input = aliasedCmd
		}
		args = strings.Fields(input)
	}

	// Check for built-in commands
	switch args[0] {
	case "cd":
		// Support 'cd' to home directory with no arguments
		if len(args) < 2 {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %v", err)
			}
			return os.Chdir(homeDir)
		}
		// Change the directory and return the error
		return os.Chdir(args[1])
	case "exit":
		os.Exit(0)
	}

	// Prepare the command to execute
	cmd := exec.Command(args[0], args[1:]...)

	// Set the correct output devices
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute '%s': %v", args[0], err)
	}
	return nil
}
