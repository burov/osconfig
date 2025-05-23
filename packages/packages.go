//  Copyright 2017 Google Inc. All Rights Reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package packages provides package management functions for Windows and Linux
// systems.
package packages

import (
	"context"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/osconfig/clog"
	"github.com/GoogleCloudPlatform/osconfig/osinfo"
	"github.com/GoogleCloudPlatform/osconfig/util"
)

var (
	// AptExists indicates whether apt is installed.
	AptExists bool
	// DpkgExists indicates whether dpkg is installed.
	DpkgExists bool
	// DpkgQueryExists indicates whether dpkg-query is installed.
	DpkgQueryExists bool
	// YumExists indicates whether yum is installed.
	YumExists bool
	// ZypperExists indicates whether zypper is installed.
	ZypperExists bool
	// RPMExists indicates whether rpm is installed.
	RPMExists bool
	// RPMQueryExists indicates whether rpmquery is installed.
	RPMQueryExists bool
	// COSPkgInfoExists indicates whether COS package information is available.
	COSPkgInfoExists bool
	// GemExists indicates whether gem is installed.
	GemExists bool
	// PipExists indicates whether pip is installed.
	PipExists bool
	// GooGetExists indicates whether googet is installed.
	GooGetExists bool
	// MSIExists indicates whether MSIs can be installed.
	MSIExists bool

	noarch = osinfo.NormalizeArchitecture("noarch")

	runner = util.CommandRunner(&util.DefaultRunner{})

	ptyrunner = util.CommandRunner(&ptyRunner{})
)

// PackageUpdatesProvider define contract to extract available updates from the VM.
type PackageUpdatesProvider interface {
	GetPackageUpdates(context.Context) (Packages, error)
}

// InstalledPackagesProvider define contract to extract installed packages from the VM.
type InstalledPackagesProvider interface {
	GetInstalledPackages(context.Context) (Packages, error)
}

type defaultUpdatesProvider struct{}

// NewPackageUpdatesProvider return fully initialize provider.
func NewPackageUpdatesProvider() PackageUpdatesProvider {
	return defaultUpdatesProvider{}
}

func (p defaultUpdatesProvider) GetPackageUpdates(ctx context.Context) (Packages, error) {
	return GetPackageUpdates(ctx)
}

type defaultInstalledPackagesProvider struct{}

// NewInstalledPackagesProvider returns fully initialized provider.
func NewInstalledPackagesProvider() InstalledPackagesProvider {
	return defaultInstalledPackagesProvider{}
}

func (p defaultInstalledPackagesProvider) GetInstalledPackages(ctx context.Context) (Packages, error) {
	return GetInstalledPackages(ctx)
}

// Packages is a selection of packages based on their manager.
type Packages struct {
	Yum                []*PkgInfo            `json:"yum,omitempty"`
	Rpm                []*PkgInfo            `json:"rpm,omitempty"`
	Apt                []*PkgInfo            `json:"apt,omitempty"`
	Deb                []*PkgInfo            `json:"deb,omitempty"`
	Zypper             []*PkgInfo            `json:"zypper,omitempty"`
	ZypperPatches      []*ZypperPatch        `json:"zypperPatches,omitempty"`
	COS                []*PkgInfo            `json:"cos,omitempty"`
	Gem                []*PkgInfo            `json:"gem,omitempty"`
	Pip                []*PkgInfo            `json:"pip,omitempty"`
	GooGet             []*PkgInfo            `json:"googet,omitempty"`
	WUA                []*WUAPackage         `json:"wua,omitempty"`
	QFE                []*QFEPackage         `json:"qfe,omitempty"`
	WindowsApplication []*WindowsApplication `json:"-"`
}

// PkgInfo describes a package.
type PkgInfo struct {
	Name, Arch, RawArch, Version string

	Source Source
}

// Source represents source package from which binary package was built.
type Source struct {
	Name, Version string
}

func (i *PkgInfo) String() string {
	return fmt.Sprintf("%s %s %s", i.Name, i.Arch, i.Version)
}

// ZypperPatch describes a Zypper patch.
type ZypperPatch struct {
	Name, Category, Severity, Summary string
}

// WUAPackage describes a Windows Update Agent package.
type WUAPackage struct {
	LastDeploymentChangeTime time.Time
	Title                    string
	Description              string
	SupportURL               string
	UpdateID                 string
	Categories               []string
	KBArticleIDs             []string
	MoreInfoURLs             []string
	CategoryIDs              []string
	RevisionNumber           int32
}

// QFEPackage describes a Windows Quick Fix Engineering package.
type QFEPackage struct {
	Caption, Description, HotFixID, InstalledOn string
}

// WindowsApplication describes a Windows Application.
type WindowsApplication struct {
	DisplayName    string
	DisplayVersion string
	InstallDate    time.Time
	Publisher      string
	HelpLink       string
}

func run(ctx context.Context, cmd string, args []string) ([]byte, error) {
	stdout, stderr, err := runner.Run(ctx, exec.CommandContext(ctx, cmd, args...))
	if err != nil {
		return nil, fmt.Errorf("error running %s with args %q: %v, stdout: %q, stderr: %q", cmd, args, err, stdout, stderr)
	}
	return stdout, nil
}

func runWithDeadline(ctx context.Context, timeout time.Duration, cmd string, args []string) ([]byte, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return run(ctxWithTimeout, cmd, args)
}

func formatFieldsMappingToFormattingString(fieldsMapping map[string]string) string {
	fieldsDescriptors := make([]string, 0, len(fieldsMapping))

	for name, selector := range fieldsMapping {
		// Format field name and its selector to one single entry separated by ":" and each of them wrapped in quotes
		// Examples:
		// name:source_name, selector:${source:Package -> ""source_name":"${source:Package}"".
		// name:source_name, selector:%{NAME} -> ""source_name":"%{NAME}"".
		fieldsDescriptors = append(fieldsDescriptors, fmt.Sprintf("\"%s\":\"%s\"", name, selector))
	}

	// Sort descriptors to get predictable result.
	sort.Strings(fieldsDescriptors)

	// Returns string to format all information in json
	// Example: {"package":"${Package}","architecture":"${Architecture}","version":"${Version}","status":"${db:Status-Status}"...}\n
	// See dpkgInfoFieldsMapping for full set of fields.
	return "\\{" + strings.Join(fieldsDescriptors, ",") + "\\}\n"
}

type packageMetadata struct {
	Package       string `json:"package"`
	Architecture  string `json:"architecture"`
	Version       string `json:"version"`
	Status        string `json:"status"`
	SourceName    string `json:"source_name"`
	SourceVersion string `json:"source_version"`
}

func pkgInfoFromPackageMetadata(pm packageMetadata) *PkgInfo {
	return &PkgInfo{
		Name:    pm.Package,
		Arch:    osinfo.NormalizeArchitecture(pm.Architecture),
		Version: pm.Version,
		Source: Source{
			Name:    pm.SourceName,
			Version: pm.SourceVersion,
		},
	}
}

type ptyRunner struct{}

func (p *ptyRunner) Run(ctx context.Context, cmd *exec.Cmd) ([]byte, []byte, error) {
	clog.Debugf(ctx, "Running %q with args %q\n", cmd.Path, cmd.Args[1:])
	stdout, stderr, err := runWithPty(cmd)
	clog.Debugf(ctx, "%s %q output:\n%s", cmd.Path, cmd.Args[1:], strings.ReplaceAll(string(stdout), "\n", "\n "))
	return stdout, stderr, err
}

// SetCommandRunner allows external clients to set a custom commandRunner.
func SetCommandRunner(commandRunner util.CommandRunner) {
	runner = commandRunner
}

// SetPtyCommandRunner allows external clients to set a custom
// custom commandRunner.
func SetPtyCommandRunner(commandRunner util.CommandRunner) {
	ptyrunner = commandRunner
}
