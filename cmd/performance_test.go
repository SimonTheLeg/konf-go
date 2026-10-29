package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
)

// BenchmarkCLI includes executable loading and package initialization, which
// in-process command benchmarks omit.
func BenchmarkCLI(b *testing.B) {
	binary := os.Getenv("KONF_BENCH_BINARY")
	if binary == "" {
		binary = filepath.Join(b.TempDir(), "konf-go")
		if output, err := exec.Command("go", "build", "-o", binary, "..").CombinedOutput(); err != nil {
			b.Fatalf("build: %v\n%s", err, output)
		}
	}
	for _, command := range [][]string{{"version"}, {"shellwrapper", "bash"}, {"completion", "bash"}} {
		b.Run(command[0], func(b *testing.B) {
			dir := b.TempDir()
			args := append([]string{"--konf-dir=" + dir, "--silent"}, command...)
			if err := exec.Command(binary, args...).Run(); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := exec.Command(binary, args...).Run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	for _, count := range []int{1, 1000} {
		b.Run(fmt.Sprintf("restore/%d", count), func(b *testing.B) {
			dir := b.TempDir()
			storeDir := filepath.Join(dir, "store")
			if err := os.MkdirAll(storeDir, 0700); err != nil {
				b.Fatal(err)
			}
			for i := 0; i < count; i++ {
				path := filepath.Join(storeDir, strconv.Itoa(i)+".yaml")
				if err := os.WriteFile(path, []byte("apiVersion: v1\nkind: Config\n"), 0600); err != nil {
					b.Fatal(err)
				}
			}
			if err := os.WriteFile(filepath.Join(dir, "latestkonf"), []byte("0"), 0600); err != nil {
				b.Fatal(err)
			}
			args := []string{"--konf-dir=" + dir, "--silent", "set", "-"}
			if err := exec.Command(binary, args...).Run(); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := exec.Command(binary, args...).Run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	for _, count := range []int{1, 100} {
		b.Run(fmt.Sprintf("cleanup/%d", count), func(b *testing.B) {
			dir := b.TempDir()
			activeDir := filepath.Join(dir, "active")
			if err := os.MkdirAll(activeDir, 0700); err != nil {
				b.Fatal(err)
			}
			args := []string{"--konf-dir=" + dir, "--silent", "cleanup"}
			if err := exec.Command(binary, args...).Run(); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				for pid := 1; pid <= count; pid++ {
					path := filepath.Join(activeDir, strconv.Itoa(2000000000+pid)+".yaml")
					if err := os.WriteFile(path, nil, 0600); err != nil {
						b.Fatal(err)
					}
				}
				b.StartTimer()
				if err := exec.Command(binary, args...).Run(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
