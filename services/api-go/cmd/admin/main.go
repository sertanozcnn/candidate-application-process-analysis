package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"capa/services/api-go/internal/config"
	"capa/services/api-go/internal/modules/adminuser"
	"capa/services/api-go/internal/platform/password"
	"capa/services/api-go/internal/platform/postgres"

	"golang.org/x/term"
)

func main() {
	if err := run(context.Background(), os.Args[1:], os.Stdin, os.Stdout); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "admin command failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, input *os.File, output io.Writer) error {
	command := "create"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command = args[0]
		args = args[1:]
	}

	if command != "create" {
		return fmt.Errorf("unknown admin command %q", command)
	}

	return create(ctx, args, input, output)
}

func create(ctx context.Context, args []string, input *os.File, output io.Writer) error {
	flags := flag.NewFlagSet("admin create", flag.ContinueOnError)
	flags.SetOutput(output)
	email := flags.String("email", "", "admin email address")

	if err := flags.Parse(args); err != nil {
		return err
	}

	normalizedEmail := adminuser.NormalizeEmail(*email)
	if normalizedEmail == "" {
		return errors.New("--email is required")
	}

	plainPassword, err := readPassword(input, output)
	if err != nil {
		return err
	}

	passwordHash, err := password.Hash(plainPassword)
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	repo := adminuser.NewRepository(db)
	if err := repo.Create(ctx, normalizedEmail, passwordHash); err != nil {
		if errors.Is(err, adminuser.ErrAlreadyExists) {
			return fmt.Errorf("admin user %q already exists", normalizedEmail)
		}

		return err
	}

	_, _ = fmt.Fprintf(output, "admin user created: %s\n", normalizedEmail)
	return nil
}

func readPassword(input *os.File, output io.Writer) (string, error) {
	_, _ = fmt.Fprint(output, "Password: ")

	var value []byte
	var err error
	if term.IsTerminal(int(input.Fd())) {
		value, err = term.ReadPassword(int(input.Fd()))
		_, _ = fmt.Fprintln(output)
	} else {
		line, readErr := bufio.NewReader(input).ReadString('\n')
		value = []byte(line)
		err = readErr
	}
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	plainPassword := strings.TrimSpace(string(value))
	if plainPassword == "" {
		return "", password.ErrEmptyPassword
	}

	return plainPassword, nil
}
