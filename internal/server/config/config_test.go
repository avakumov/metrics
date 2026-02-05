package config

import (
	"flag"
	"os"
	"strings"
	"testing"
)

func TestGetOptions_Defaults(t *testing.T) {
	// Сохраняем оригинальные значения
	origArgs := os.Args
	origEnv := os.Environ()
	defer func() {
		os.Args = origArgs
		os.Clearenv()
		for _, e := range origEnv {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				os.Setenv(parts[0], parts[1])
			}
		}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Подготавливаем окружение
	os.Args = []string{"test"}
	os.Clearenv()

	// Получаем опции
	options := GetOptions()

	// Проверяем значения по умолчанию
	if options.Address != "localhost:8080" {
		t.Errorf("expected Address 'localhost:8080', got %s", options.Address)
	}
}

func TestGetOptions_FromFlags(t *testing.T) {
	// Сохраняем и восстанавливаем
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Устанавливаем флаги
	os.Args = []string{
		"test",
		"-a", "example.com:9090",
	}

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	options := GetOptions()

	// Проверяем

	if options.Address != "example.com:9090" {
		t.Errorf("expected Address 'example.com:9090', got %s", options.Address)
	}
}

func TestGetOptions_FromEnv(t *testing.T) {
	// Сохраняем окружение
	origArgs := os.Args
	origEnv := map[string]string{
		"ADDRESS": os.Getenv("ADDRESS"),
	}
	defer func() {
		os.Args = origArgs
		for k, v := range origEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Настраиваем тест
	os.Args = []string{"test"}
	os.Setenv("ADDRESS", "example.com:8081")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	options := GetOptions()

	// Проверяем, что env переопределило defaults
	if options.Address != "example.com:8081" {
		t.Errorf("expected address example.com:8081 from env, got %s", options.Address)
	}
}

func TestGetOptions_FlagsOverrideEnv(t *testing.T) {
	origArgs := os.Args
	origEnv := map[string]string{
		"ADDRESS": os.Getenv("ADDRESS"),
	}
	defer func() {
		os.Args = origArgs
		for k, v := range origEnv {
			if v != "" {
				os.Setenv(k, v)
			} else {
				os.Unsetenv(k)
			}
		}
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	}()

	// Устанавливаем env и флаги
	os.Args = []string{"test", "-a", "args.com:9091"}
	os.Setenv("ADDRESS", "exe.com:8088")

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	options := GetOptions()

	// Флаги должны переопределить env

	if options.Address != "args.com:9091" {
		t.Errorf("expected address args.com:9091 from flag -a, got %s", options.Address)
	}
}
