package analysis

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// Synthetic source fixtures
// ---------------------------------------------------------------------------

// goFunc generates a Go function with n meaningful body lines (assignments,
// branches, loops, calls) so the flow extractor actually classifies steps.
func goFunc(name string, n int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "func %s(input int) int {\n", name)
	for i := 0; i < n; i++ {
		switch i % 5 {
		case 0:
			fmt.Fprintf(&b, "\tx%d := input + %d\n", i, i)
		case 1:
			fmt.Fprintf(&b, "\tif x%d > 0 {\n\t\tx%d++\n\t}\n", i-1, i-1)
		case 2:
			fmt.Fprintf(&b, "\tfor j := 0; j < x%d; j++ {\n\t\tx%d += j\n\t}\n", i-1, i-1)
		case 3:
			fmt.Fprintf(&b, "\thelper(x%d)\n", i-1)
		case 4:
			fmt.Fprintf(&b, "\tresult%d := compute(x%d, %d)\n", i, i-1, i)
		}
	}
	b.WriteString("\treturn 0\n}\n")
	return b.String()
}

// goFile builds a complete Go file with one large function and a helper stub.
func goFile(funcLines int) string {
	return "package bench\n\nfunc helper(v int) {}\nfunc compute(a, b int) int { return a + b }\n\n" +
		goFunc("target", funcLines)
}

// ---------------------------------------------------------------------------
// extractFlow benchmarks — pure parsing, no I/O
// ---------------------------------------------------------------------------

func BenchmarkExtractFlow_Small(b *testing.B) {
	// ~10 body lines → roughly what a typical short handler does
	content := goFile(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractFlow(content, "bench.go", "target")
	}
}

func BenchmarkExtractFlow_Medium(b *testing.B) {
	// ~50 body lines — a realistic mid-size function
	content := goFile(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractFlow(content, "bench.go", "target")
	}
}

func BenchmarkExtractFlow_Large(b *testing.B) {
	// ~200 body lines — a large function, stress case
	content := goFile(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractFlow(content, "bench.go", "target")
	}
}

func BenchmarkExtractFlow_Huge(b *testing.B) {
	// ~500 body lines — deliberately extreme
	content := goFile(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = extractFlow(content, "bench.go", "target")
	}
}

// ---------------------------------------------------------------------------
// buildCFG benchmarks — graph construction from pre-parsed steps
// ---------------------------------------------------------------------------

func BenchmarkBuildCFG_NoLimit(b *testing.B) {
	content := goFile(100)
	steps, err := extractFlow(content, "bench.go", "target")
	if err != nil || len(steps) == 0 {
		b.Fatalf("extractFlow failed: %v (steps=%d)", err, len(steps))
	}
	b.ReportMetric(float64(len(steps)), "steps")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// No truncation limit — measure raw CFG construction cost
		buildCFG("bench.go", "target", 1, steps, len(steps)+1, false)
	}
}

func BenchmarkBuildCFG_WithLimit(b *testing.B) {
	content := goFile(100)
	steps, _ := extractFlow(content, "bench.go", "target")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildCFG("bench.go", "target", 1, steps, MaxRootNodes, true)
	}
}

func BenchmarkBuildCFG_HighLimit(b *testing.B) {
	content := goFile(100)
	steps, _ := extractFlow(content, "bench.go", "target")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// What a raised limit (e.g. 50 nodes) would cost
		buildCFG("bench.go", "target", 1, steps, 50, false)
	}
}

// ---------------------------------------------------------------------------
// End-to-end Root() benchmarks — includes file I/O via temp file
// ---------------------------------------------------------------------------

func writeBenchFile(b *testing.B, root, name, content string) {
	b.Helper()
	path := root + "/" + name
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		b.Fatal(err)
	}
}

func BenchmarkRoot_Small(b *testing.B) {
	root := b.TempDir()
	writeBenchFile(b, root, "bench.go", goFile(10))
	builder, _ := New(root, "rg")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Root(context.Background(), "bench.go", "target")
	}
}

func BenchmarkRoot_Medium(b *testing.B) {
	root := b.TempDir()
	writeBenchFile(b, root, "bench.go", goFile(50))
	builder, _ := New(root, "rg")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Root(context.Background(), "bench.go", "target")
	}
}

func BenchmarkRoot_Large(b *testing.B) {
	root := b.TempDir()
	writeBenchFile(b, root, "bench.go", goFile(200))
	builder, _ := New(root, "rg")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = builder.Root(context.Background(), "bench.go", "target")
	}
}

// ---------------------------------------------------------------------------
// Concurrent load simulation — 100 goroutines firing simultaneously
// ---------------------------------------------------------------------------

func BenchmarkRoot_Concurrent100_Medium(b *testing.B) {
	root := b.TempDir()
	writeBenchFile(b, root, "bench.go", goFile(50))
	builder, _ := New(root, "rg")
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = builder.Root(context.Background(), "bench.go", "target")
		}
	})
}

// Explicit 100-goroutine burst — closest to "100 visitors at once"
func BenchmarkRoot_100GoroutineBurst(b *testing.B) {
	root := b.TempDir()
	writeBenchFile(b, root, "bench.go", goFile(50))
	builder, _ := New(root, "rg")
	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(100)
		for g := 0; g < 100; g++ {
			go func() {
				defer wg.Done()
				_, _ = builder.Root(ctx, "bench.go", "target")
			}()
		}
		wg.Wait()
	}
}

// ---------------------------------------------------------------------------
// Limit sensitivity — how much does raising MaxRootNodes actually cost?
// ---------------------------------------------------------------------------

func BenchmarkLimitSensitivity(b *testing.B) {
	content := goFile(200)
	steps, _ := extractFlow(content, "bench.go", "target")
	total := len(steps)

	for _, limit := range []int{8, 16, 25, 50, 100, total} {
		limit := limit
		label := fmt.Sprintf("limit=%d", limit)
		if limit == total {
			label = "unlimited"
		}
		b.Run(label, func(b *testing.B) {
			b.ReportMetric(float64(limit), "limit")
			for i := 0; i < b.N; i++ {
				buildCFG("bench.go", "target", 1, steps, limit, limit < total)
			}
		})
	}
}
