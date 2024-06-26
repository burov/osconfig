//  Copyright 2020 Google Inc. All Rights Reserved.
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

//go:build linux && (386 || amd64)
// +build linux
// +build 386 amd64

package packages

import (
	"errors"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"cos.googlesource.com/cos/tools.git/src/pkg/cos"
)

func TestParseInstalledCOSPackages(t *testing.T) {
	readMachineArch = func() (string, error) {
		return "", errors.New("failed to obtain machine architecture")
	}
	if _, err := parseInstalledCOSPackages(&cos.PackageInfo{}); err == nil {
		t.Errorf("did not get expected error")
	}

	readMachineArch = func() (string, error) {
		return "x86_64", nil
	}

	pkg0 := cos.Package{Category: "dev-util", Name: "foo-x", Version: "1.2.3", EbuildVersion: "someversion"}
	expect0 := &PkgInfo{Name: "dev-util/foo-x", Arch: "x86_64", Version: "1.2.3"}
	pkg1 := cos.Package{Category: "app-admin", Name: "bar", Version: "0.1"}
	expect1 := &PkgInfo{Name: "app-admin/bar", Arch: "x86_64", Version: "0.1"}

	pkgInfo := &cos.PackageInfo{InstalledPackages: []cos.Package{pkg0, pkg1}}
	parsed, err := parseInstalledCOSPackages(pkgInfo)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(parsed[0], expect0) {
		t.Errorf("parseInstalledCOSPackages pkg0: %v, want: %v", parsed[0], expect0)
	}
	if !reflect.DeepEqual(parsed[1], expect1) {
		t.Errorf("parseInstalledCOSPackages pkg1: %v, want: %v", parsed[1], expect1)
	}
}

func TestInstalledCOSPackages(t *testing.T) {
	testDataJSON := `{
    "installedPackages": [
        {
            "category": "app-arch",
            "name": "gzip",
            "version": "1.9",
			"ebuildverison": "someotherversion"
        },
        {
            "category": "dev-libs",
            "name": "popt",
            "version": "1.16"
        },
        {
            "category": "app-emulation",
            "name": "docker-credential-helpers",
            "version": "0.6.3"
        },
        {
            "category": "_not.real-category1+",
            "name": "_not-real_package1",
            "version": "12.34.56.78"
        },
        {
            "category": "_not.real-category1+",
            "name": "_not-real_package2",
            "version": "12.34.56.78"
        },
        {
            "category": "_not.real-category1+",
            "name": "_not-real_package3",
            "version": "12.34.56.78_rc3"
        },
        {
            "category": "_not.real-category1+",
            "name": "_not-real_package4",
            "version": "12.34.56.78_rc3"
        },
        {
            "category": "_not.real-category1+",
            "name": "_not-real_package5",
            "version": "12.34.56.78_pre2_rc3"
        },
        {
            "category": "_not.real-category2+",
            "name": "_not-real_package1",
            "version": "12.34.56.78q"
        },
        {
            "category": "_not.real-category2+",
            "name": "_not-real_package2",
            "version": "12.34.56.78q"
        },
        {
            "category": "_not.real-category2+",
            "name": "_not-real_package3",
            "version": "12.34.56.78q_rc3"
        },
        {
            "category": "_not.real-category2+",
            "name": "_not-real_package4",
            "version": "12.34.56.78q_rc3"
        },
        {
            "category": "_not.real-category2+",
            "name": "_not-real_package5",
            "version": "12.34.56.78q_pre2_rc3"
        }
    ]
}`

	testFile, err := ioutil.TempFile("", "cos_pkg_info_test")
	if err != nil {
		t.Fatalf("Failed to create tempfile: %v", err)
	}
	defer os.Remove(testFile.Name())
	_, err = testFile.WriteString(testDataJSON)
	if err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}
	err = testFile.Close()
	if err != nil {
		t.Fatalf("Failed to close test file: %v", err)
	}

	expected := []*PkgInfo{
		{Name: "app-arch/gzip", Arch: "x86_64", Version: "1.9"},
		{Name: "dev-libs/popt", Arch: "x86_64", Version: "1.16"},
		{Name: "app-emulation/docker-credential-helpers", Arch: "x86_64", Version: "0.6.3"},
		{Name: "_not.real-category1+/_not-real_package1", Arch: "x86_64", Version: "12.34.56.78"},
		{Name: "_not.real-category1+/_not-real_package2", Arch: "x86_64", Version: "12.34.56.78"},
		{Name: "_not.real-category1+/_not-real_package3", Arch: "x86_64", Version: "12.34.56.78_rc3"},
		{Name: "_not.real-category1+/_not-real_package4", Arch: "x86_64", Version: "12.34.56.78_rc3"},
		{Name: "_not.real-category1+/_not-real_package5", Arch: "x86_64", Version: "12.34.56.78_pre2_rc3"},
		{Name: "_not.real-category2+/_not-real_package1", Arch: "x86_64", Version: "12.34.56.78q"},
		{Name: "_not.real-category2+/_not-real_package2", Arch: "x86_64", Version: "12.34.56.78q"},
		{Name: "_not.real-category2+/_not-real_package3", Arch: "x86_64", Version: "12.34.56.78q_rc3"},
		{Name: "_not.real-category2+/_not-real_package4", Arch: "x86_64", Version: "12.34.56.78q_rc3"},
		{Name: "_not.real-category2+/_not-real_package5", Arch: "x86_64", Version: "12.34.56.78q_pre2_rc3"},
	}

	readMachineArch = func() (string, error) {
		return "", errors.New("failed to obtain machine architecture")
	}
	readCOSPackageInfo = func() (*cos.PackageInfo, error) {
		info, err := cos.GetPackageInfoFromFile(testFile.Name())
		return &info, err
	}
	if _, err := InstalledCOSPackages(); err == nil {
		t.Errorf("did not get expected error from readMachineArch")
	}

	readMachineArch = func() (string, error) {
		return "x86_64", nil
	}
	readCOSPackageInfo = func() (*cos.PackageInfo, error) {
		info, err := cos.GetPackageInfoFromFile("_" + testFile.Name())
		return &info, err
	}
	if _, err := InstalledCOSPackages(); err == nil {
		t.Errorf("did not get expected error fro readCOSPackageInfo")
	}

	readMachineArch = func() (string, error) {
		return "x86_64", nil
	}
	readCOSPackageInfo = func() (*cos.PackageInfo, error) {
		info, err := cos.GetPackageInfoFromFile(testFile.Name())
		return &info, err
	}
	ret, err := InstalledCOSPackages()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(ret) != len(expected) {
		t.Errorf("Length is wrong. want: %d, got: %d",
			len(expected), len(ret))
	}
	if !reflect.DeepEqual(ret, expected) {
		t.Errorf("InstalledCOSPackages() returned: %v, want: %v", ret, expected)
	}

}
