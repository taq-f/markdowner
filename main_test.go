package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParsePathSuccess(t *testing.T) {

	type TestCase struct {
		input             string
		output            string
		expectedInputPath string
		expectedBasePath  string
		expectedOutPath   string
	}

	curAbsPath, _ := os.Getwd()

	// create directories and files for this test to run
	sampleDir := filepath.Join(curAbsPath, "sample")
	os.MkdirAll(sampleDir, os.ModeDir)
	file, err := os.Create(filepath.Join(sampleDir, "sample.md"))
	if err != nil {
		t.Fatal("failed to create sample.md. can't continue.")
	}

	testCases := []TestCase{
		TestCase{
			"sample",
			"dist",
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "dist"),
		},
		TestCase{
			"sample",
			"",
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "sample"),
		},
		TestCase{
			"sample\\sample.md",
			"dist\\sample.html",
			filepath.Join(curAbsPath, "sample", "sample.md"),
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "dist", "sample.html"),
		},
		TestCase{
			"",
			"dist\\dist_sub",
			curAbsPath,
			curAbsPath,
			filepath.Join(curAbsPath, "dist", "dist_sub"),
		},
		TestCase{
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "dist", "dist_sub"),
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "sample"),
			filepath.Join(curAbsPath, "dist", "dist_sub"),
		},
	}

	for i, testCase := range testCases {
		inputPath, basePath, outPath, err := parsePath(testCase.input, testCase.output)

		if inputPath != testCase.expectedInputPath {
			t.Errorf("\n%d input path error\ngot %v\nwant %v", i, inputPath, testCase.expectedInputPath)
		}
		if basePath != testCase.expectedBasePath {
			t.Errorf("\n%d base path error\ngot %v\nwant %v", i, basePath, testCase.expectedBasePath)
		}
		if outPath != testCase.expectedOutPath {
			t.Errorf("\n%d out path error\ngot %v\nwant %v", i, outPath, testCase.expectedOutPath)
		}
		if err != nil {
			t.Errorf("\n%d unexpectedly error occurs: %v", i, err)
		}
	}

	file.Close()
	if err := os.RemoveAll(sampleDir); err != nil {
		t.Fatalf("failed to remove sample directory: %v", err)
	}
}

func TestParsePathFail(t *testing.T) {
	// input dir not found
	_, _, _, err := parsePath("shouldnotexists", "")
	if err == nil {
		t.Error("getPath unexpectedly did not throw error while input directory does not exist.")
	}
}

func TestIsDirTrue(t *testing.T) {
	curPath, _ := os.Getwd()
	testDirPath := filepath.Join(curPath, "test_assets")
	if !isDir(testDirPath) {
		t.Errorf("isDir returned false while the directory exists: %v", testDirPath)
	}
}

func TestIsDirFalse(t *testing.T) {
	curPath, _ := os.Getwd()
	testDirPath := filepath.Join(curPath, "test_assets", "path_not_exists")
	if isDir(testDirPath) {
		t.Errorf("isDir returned true while the directory does not exist: %v", testDirPath)
	}

	testDirPath = filepath.Join(curPath, "test_assets", "sample.html")
	if isDir(testDirPath) {
		t.Errorf("isDir returned true while the path points to a file: %v", testDirPath)
	}
}
