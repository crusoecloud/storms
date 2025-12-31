package utils

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

const (
	requiredNote = "(required)"
)

// FlagBuilder holds the command to register flags on.
type FlagBuilder struct {
	cmd *cobra.Command
}

// NewFlagBuilder returns a builder for the given command.
func NewFlagBuilder(cmd *cobra.Command) *FlagBuilder {
	return &FlagBuilder{cmd: cmd}
}

// String registers a string flag with optional required setting.
func (b *FlagBuilder) String(name, shorthand, usage string, required bool) *FlagBuilder {
	if required && !strings.Contains(usage, requiredNote) {
		usage = strings.TrimSpace(usage) + " " + requiredNote
	}
	b.cmd.Flags().StringP(name, shorthand, "", usage)
	if required {
		_ = b.cmd.MarkFlagRequired(name) //nolint:errcheck // this should not fail
	}

	return b
}

func (b *FlagBuilder) StringCSV(name, shorthand, usage string, required bool) *FlagBuilder {
	if required && !strings.Contains(usage, requiredNote) {
		usage = strings.TrimSpace(usage) + " " + requiredNote
	}

	// Register the flag as a string
	b.cmd.Flags().StringP(name, shorthand, "", usage)
	if required {
		_ = b.cmd.MarkFlagRequired(name) //nolint:errcheck // will no fail because we guarantee the flag exists
	}

	// Wrap RunE to parse CSV into a string slice before original RunE
	originalRunE := b.cmd.RunE
	b.cmd.RunE = func(cmd *cobra.Command, args []string) error {
		val, err := cmd.Flags().GetString(name)
		if err != nil {
			return err //nolint:wrapcheck // allow this because function is anonymous
		}

		if val != "" {
			// Save the parsed CSV as a StringSlice flag for later retrieval
			slice := splitAndTrim(val)
			_ = cmd.Flags().Set(name, strings.Join(slice, ",")) //nolint:errcheck // flag will be valid
		}

		if originalRunE != nil {
			return originalRunE(cmd, args)
		}

		return nil
	}

	return b
}

// Helper to split and trim CSV values.
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts
}

// Int registers an int flag.
func (b *FlagBuilder) Int(ptr *int, name, shorthand, usage string, required bool) *FlagBuilder {
	if required && !strings.Contains(usage, requiredNote) {
		usage = strings.TrimSpace(usage) + " " + requiredNote
	}
	b.cmd.Flags().IntVarP(ptr, name, shorthand, 0, usage)
	if required {
		_ = b.cmd.MarkFlagRequired(name) //nolint:errcheck // this should not fail
	}

	return b
}

// Int registers an int flag.
func (b *FlagBuilder) Uint(name, shorthand, usage string, required bool) *FlagBuilder {
	if required && !strings.Contains(usage, requiredNote) {
		usage = strings.TrimSpace(usage) + " " + requiredNote
	}
	b.cmd.Flags().UintP(name, shorthand, 0, usage)
	if required {
		_ = b.cmd.MarkFlagRequired(name) //nolint:errcheck // this should not fail
	}

	return b
}

// Bool registers a bool flag.
func (b *FlagBuilder) Bool(name, shorthand, usage string, required bool) *FlagBuilder {
	if required && !strings.Contains(usage, requiredNote) {
		usage = strings.TrimSpace(usage) + " " + requiredNote
	}
	b.cmd.Flags().BoolP(name, shorthand, false, usage)
	if required {
		_ = b.cmd.MarkFlagRequired(name) //nolint:errcheck // this should not fail
	}

	return b
}

// StringToString registers a map of strings flag.
// Input format: --flag key1=value1,key2=value2.
func (b *FlagBuilder) StringToString(name, shorthand, usage string, required bool) *FlagBuilder {
	if required && !strings.Contains(usage, requiredNote) {
		usage = strings.TrimSpace(usage) + " " + requiredNote
	}
	// StringToStringP registers a flag that parses "k=v,k2=v2" into map[string]string
	b.cmd.Flags().StringToStringP(name, shorthand, nil, usage)
	if required {
		_ = b.cmd.MarkFlagRequired(name) //nolint:errcheck // this should not fail
	}

	return b
}

func MustGetStringFlag(cmd *cobra.Command, name string) string {
	val, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(fmt.Sprintf("flag %q not defined: %v", name, err))
	}

	return val
}

// MustGetStringCSVFlag fetches a CSV string flag and returns it as a slice of strings.
// Panics if the flag does not exist.
func MustGetStringCSVFlag(cmd *cobra.Command, name string) []string {
	val, err := cmd.Flags().GetString(name)
	if err != nil {
		panic(fmt.Sprintf("flag %q not defined: %v", name, err))
	}

	if val == "" {
		return nil
	}

	return splitAndTrim(val)
}

func MustGetUintFlag(cmd *cobra.Command, name string) uint {
	val, err := cmd.Flags().GetUint(name)
	if err != nil {
		panic(fmt.Sprintf("flag %q not defined: %v", name, err))
	}

	return val
}

func MustGetBoolFlag(cmd *cobra.Command, name string) bool {
	val, err := cmd.Flags().GetBool(name)
	if err != nil {
		panic(fmt.Sprintf("flag %q not defined: %v", name, err))
	}

	return val
}

// MustGetStringToString retrieves a map[string]string flag or panics if it doesn't exist.
func MustGetStringToStringFlag(cmd *cobra.Command, name string) map[string]string {
	val, err := cmd.Flags().GetStringToString(name)
	if err != nil {
		panic(fmt.Sprintf("flag %q not defined or invalid: %v", name, err))
	}

	return val
}
