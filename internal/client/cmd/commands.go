package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/Himany/GophKeeper/internal/client"
	"github.com/Himany/GophKeeper/internal/client/storage"
	"github.com/Himany/GophKeeper/internal/config"
	"github.com/Himany/GophKeeper/pkg/api"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	cfg           *config.ClientConfig
	clientStorage *storage.LocalStorage
	apiClient     *client.Client
)

var rootCmd = &cobra.Command{
	Use:   "gophkeeper",
	Short: "GophKeeper - secure password manager",
	Long: `GophKeeper is a secure client-server password manager that allows you to 
safely store login credentials, text data, binary files, and credit card information.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gophkeeper.yaml)")
	rootCmd.PersistentFlags().String("server", "http://localhost:8080", "server URL")

	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(versionCmd)
}

var cfgFile string

func initConfig() {
	var err error
	cfg, err = config.LoadClientConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	if serverURL := rootCmd.Flag("server").Value.String(); serverURL != "" {
		cfg.ServerURL = serverURL
	}

	clientStorage, err = storage.NewLocalStorage(cfg)
	if err != nil {
		fmt.Printf("Error initializing local storage: %v\n", err)
		os.Exit(1)
	}

	apiClient = client.NewClient(cfg.ServerURL)

	if token, _ := clientStorage.LoadToken(); token != "" {
		apiClient.SetToken(token)
	}
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if username == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		if password == "" {
			fmt.Print("Password: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Printf("Error reading password: %v\n", err)
				return
			}
			password = string(bytePassword)
			fmt.Println()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := apiClient.Register(ctx, username, password)
		if err != nil {
			fmt.Printf("Registration failed: %v\n", err)
			return
		}

		if err := clientStorage.SaveToken(response.Token); err != nil {
			fmt.Printf("Warning: failed to save token: %v\n", err)
		}

		fmt.Printf("Registration successful! Welcome, %s\n", response.Username)
	},
}

func init() {
	registerCmd.Flags().StringP("username", "u", "", "Username for registration")
	registerCmd.Flags().StringP("password", "p", "", "Password for registration")
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the server",
	Run: func(cmd *cobra.Command, args []string) {
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")

		if username == "" {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		if password == "" {
			fmt.Print("Password: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				fmt.Printf("Error reading password: %v\n", err)
				return
			}
			password = string(bytePassword)
			fmt.Println()
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := apiClient.Login(ctx, username, password)
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			return
		}

		if err := clientStorage.SaveToken(response.Token); err != nil {
			fmt.Printf("Warning: failed to save token: %v\n", err)
		}

		fmt.Printf("Login successful! Welcome back, %s\n", response.Username)
	},
}

func init() {
	loginCmd.Flags().StringP("username", "u", "", "Username for login")
	loginCmd.Flags().StringP("password", "p", "", "Password for login")
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from the server",
	Run: func(cmd *cobra.Command, args []string) {
		if err := clientStorage.RemoveToken(); err != nil {
			fmt.Printf("Error during logout: %v\n", err)
			return
		}

		apiClient.SetToken("")
		fmt.Println("Logged out successfully")
	},
}

var addCmd = &cobra.Command{
	Use:   "add [type]",
	Short: "Add new data entry",
	Long: `Add a new data entry. Supported types:
  credentials - login/password pairs
  text       - arbitrary text data
  binary     - binary files
  card       - credit card information`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entryType := args[0]
		name, _ := cmd.Flags().GetString("name")
		metadata, _ := cmd.Flags().GetString("metadata")

		if name == "" {
			fmt.Print("Entry name: ")
			fmt.Scanln(&name)
		}

		var data map[string]string
		var err error

		switch entryType {
		case "credentials":
			data, err = promptCredentials()
		case "text":
			data, err = promptText()
		case "binary":
			data, err = promptBinary(cmd)
		case "card":
			data, err = promptCard()
		default:
			fmt.Printf("Unsupported entry type: %s\n", entryType)
			return
		}

		if err != nil {
			fmt.Printf("Error collecting data: %v\n", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := apiClient.CreateEntry(ctx, name, entryType, data, metadata)
		if err != nil {
			fmt.Printf("Failed to create entry: %v\n", err)
			return
		}

		fmt.Printf("Entry '%s' created successfully (ID: %s)\n", response.Name, response.ID)
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "Entry name")
	addCmd.Flags().StringP("metadata", "m", "", "Entry metadata")
	addCmd.Flags().StringP("file", "f", "", "File path for binary data")
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all data entries",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := apiClient.ListEntries(ctx)
		if err != nil {
			fmt.Printf("Failed to list entries: %v\n", err)
			return
		}

		if len(response.Entries) == 0 {
			fmt.Println("No entries found")
			return
		}

		fmt.Printf("Found %d entries:\n\n", len(response.Entries))
		for _, entry := range response.Entries {
			fmt.Printf("ID: %s\n", entry.ID)
			fmt.Printf("Name: %s\n", entry.Name)
			fmt.Printf("Type: %s\n", entry.Type)
			if entry.Metadata != "" {
				fmt.Printf("Metadata: %s\n", entry.Metadata)
			}
			fmt.Printf("Created: %s\n", time.Unix(entry.CreatedAt, 0).Format(time.RFC3339))
			fmt.Printf("Updated: %s\n", time.Unix(entry.UpdatedAt, 0).Format(time.RFC3339))
			fmt.Println("---")
		}
	},
}

var getCmd = &cobra.Command{
	Use:   "get [entry-id]",
	Short: "Get data entry by ID",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entryID, err := uuid.Parse(args[0])
		if err != nil {
			fmt.Printf("Invalid entry ID: %v\n", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		response, err := apiClient.GetEntry(ctx, entryID)
		if err != nil {
			fmt.Printf("Failed to get entry: %v\n", err)
			return
		}

		displayEntry(response)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GophKeeper Client\n")
		fmt.Printf("Version: %s\n", version)
		fmt.Printf("Build time: %s\n", buildTime)
	},
}

var (
	version   = "dev"
	buildTime = "unknown"
)

func SetVersionInfo(v, bt string) {
	version = v
	buildTime = bt
}

// Вспомогательные функции

func promptCredentials() (map[string]string, error) {
	data := make(map[string]string)
	var input string

	fmt.Print("Login: ")
	fmt.Scanln(&input)
	data["login"] = input

	fmt.Print("Password: ")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return nil, err
	}
	data["password"] = string(bytePassword)
	fmt.Println()

	fmt.Print("URL (optional): ")
	fmt.Scanln(&input)
	data["url"] = input

	return data, nil
}

func promptText() (map[string]string, error) {
	data := make(map[string]string)
	var input string

	fmt.Print("Text content: ")
	fmt.Scanln(&input)
	data["content"] = input

	return data, nil
}

func promptBinary(cmd *cobra.Command) (map[string]string, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath == "" {
		fmt.Print("File path: ")
		fmt.Scanln(&filePath)
	}

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	data := map[string]string{
		"filename": filePath,
		"content":  string(fileData),
	}

	return data, nil
}

func promptCard() (map[string]string, error) {
	data := make(map[string]string)
	var input string

	fmt.Print("Card number: ")
	fmt.Scanln(&input)
	data["number"] = input

	fmt.Print("Cardholder name: ")
	fmt.Scanln(&input)
	data["holder"] = input

	fmt.Print("Expiry date (MM/YY): ")
	fmt.Scanln(&input)
	data["expiry_date"] = input

	fmt.Print("CVV: ")
	fmt.Scanln(&input)
	data["cvv"] = input

	fmt.Print("Bank (optional): ")
	fmt.Scanln(&input)
	data["bank"] = input

	return data, nil
}

func displayEntry(entry *api.EntryResponse) {
	fmt.Printf("ID: %s\n", entry.ID)
	fmt.Printf("Name: %s\n", entry.Name)
	fmt.Printf("Type: %s\n", entry.Type)

	fmt.Println("Data:")
	for key, value := range entry.Data {
		if strings.Contains(strings.ToLower(key), "password") || strings.Contains(strings.ToLower(key), "cvv") {
			fmt.Printf("  %s: [HIDDEN]\n", key)
		} else {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	if entry.Metadata != "" {
		fmt.Printf("Metadata: %s\n", entry.Metadata)
	}

	fmt.Printf("Created: %s\n", time.Unix(entry.CreatedAt, 0).Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", time.Unix(entry.UpdatedAt, 0).Format(time.RFC3339))
}

var updateCmd = &cobra.Command{
	Use:   "update [entry-id]",
	Short: "Update data entry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Update command is not implemented yet")
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [entry-id]",
	Short: "Delete data entry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		entryID, err := uuid.Parse(args[0])
		if err != nil {
			fmt.Printf("Invalid entry ID: %v\n", err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := apiClient.DeleteEntry(ctx, entryID); err != nil {
			fmt.Printf("Failed to delete entry: %v\n", err)
			return
		}

		fmt.Println("Entry deleted successfully")
	},
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize data with server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Sync command is not implemented yet")
	},
}
