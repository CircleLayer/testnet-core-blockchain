// Copyright 2021 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
// Enhanced blockchain implementation by Circle Layer <https://circlelayer.com>

package build

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type GoToolchain struct {
	Root string // GOROOT

	// Cross-compilation variables. These are set when running the go tool.
	GOARCH string
	GOOS   string
	CC     string
}

// Go creates an invocation of the go command.
func (g *GoToolchain) Go(command string, args ...string) *exec.Cmd {
	tool := g.goTool(command, args...)

	// Configure environment for cross build.
	if g.GOARCH != "" && g.GOARCH != runtime.GOARCH {
		tool.Env = append(tool.Env, "CGO_ENABLED=1")
		tool.Env = append(tool.Env, "GOARCH="+g.GOARCH)
	}
	if g.GOOS != "" && g.GOOS != runtime.GOOS {
		tool.Env = append(tool.Env, "GOOS="+g.GOOS)
	}
	// Configure C compiler.
	if g.CC != "" {
		tool.Env = append(tool.Env, "CC="+g.CC)
	} else if os.Getenv("CC") != "" {
		tool.Env = append(tool.Env, "CC="+os.Getenv("CC"))
	}
	return tool
}

// Install creates an invocation of 'go install'. The command is configured to output
// executables to the given 'gobin' directory.
//
// This can be used to install auxiliary build tools without modifying the local go.mod and
// go.sum files. To install tools which are not required by go.mod, ensure that all module
// paths in 'args' contain a module version suffix (e.g. "...@latest").
func (g *GoToolchain) Install(gobin string, args ...string) *exec.Cmd {
	if !filepath.IsAbs(gobin) {
		panic("GOBIN must be an absolute path")
	}
	tool := g.goTool("install")
	tool.Env = append(tool.Env, "GOBIN="+gobin)
	tool.Args = append(tool.Args, "-mod=readonly")
	tool.Args = append(tool.Args, args...)

	// Ensure GOPATH is set because go install seems to absolutely require it. This uses
	// 'go env' because it resolves the default value when GOPATH is not set in the
	// environment. Ignore errors running go env and leave any complaining about GOPATH to
	// the install command.
	pathTool := g.goTool("env", "GOPATH")
	output, _ := pathTool.Output()
	tool.Env = append(tool.Env, "GOPATH="+string(output))
	return tool
}

func (g *GoToolchain) goTool(command string, args ...string) *exec.Cmd {
	if g.Root == "" {
		g.Root = runtime.GOROOT()
	}
	tool := exec.Command(filepath.Join(g.Root, "bin", "go"), command)
	tool.Args = append(tool.Args, args...)
	tool.Env = append(tool.Env, "GOROOT="+g.Root)

	// Forward environment variables to the tool, but skip compiler target settings.
	// TODO: what about GOARM?
	skip := map[string]struct{}{"GOROOT": {}, "GOARCH": {}, "GOOS": {}, "GOBIN": {}, "CC": {}}
	for _, e := range os.Environ() {
		if i := strings.IndexByte(e, '='); i >= 0 {
			if _, ok := skip[e[:i]]; ok {
				continue
			}
		}
		tool.Env = append(tool.Env, e)
	}
	return tool
}

// DownloadGo downloads the Go binary distribution and unpacks it into a temporary
// directory. It returns the GOROOT of the unpacked toolchain.
func DownloadGo(csdb *ChecksumDB, version string) string {
	// Shortcut: if the Go version that runs this script matches the
	// requested version exactly, there is no need to download anything.
	activeGo := strings.TrimPrefix(runtime.Version(), "go")
	if activeGo == version {
		log.Printf("-dlgo version matches active Go version %s, skipping download.", activeGo)
		return runtime.GOROOT()
	}

	ucache, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	// For Arm architecture, GOARCH includes ISA version.
	os := runtime.GOOS
	arch := runtime.GOARCH
	if arch == "arm" {
		arch = "armv6l"
	}
	file := fmt.Sprintf("go%s.%s-%s", version, os, arch)
	if os == "windows" {
		file += ".zip"
	} else {
		file += ".tar.gz"
	}
	url := "https://golang.org/dl/" + file
	dst := filepath.Join(ucache, file)
	if err := csdb.DownloadFile(url, dst); err != nil {
		log.Fatal(err)
	}

	godir := filepath.Join(ucache, fmt.Sprintf("geth-go-%s-%s-%s", version, os, arch))
	if err := ExtractArchive(dst, godir); err != nil {
		log.Fatal(err)
	}
	goroot, err := filepath.Abs(filepath.Join(godir, "go"))
	if err != nil {
		log.Fatal(err)
	}
	return goroot
}
