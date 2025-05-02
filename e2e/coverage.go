package e2e

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func Coverage() env.Func {
	return func(ctx context.Context, cfg *envconf.Config) (context.Context, error) {
		filePath := "coverage.out"
		// Controlla se il file esiste
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return ctx, nil
		}

		cmd := exec.Command("go", "tool", "cover", "-func="+filePath)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return ctx, err
		}

		scanner := bufio.NewScanner(bytes.NewReader(out))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "coverage:") {
				coverageLine := strings.Replace(line, "total:", "coverage:", 1) // Riformatta l'output
				fmt.Println(coverageLine)
				break
			}
		}

		return ctx, nil
	}
}
