package renderer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestRender(t *testing.T) {
	curPath, _ := os.Getwd()
	mdPath := filepath.Join(curPath, "..", "test_assets", "sample.md")
	basePath := filepath.Join(curPath, "..", "test_assets")
	outDir := filepath.Join(curPath, "..", "test_assets")
	outFilePath := filepath.Join(outDir, "sample.html")

	// cleaning up the output file
	if _, err := os.Stat(outFilePath); err == nil {
		if err := os.Remove(outFilePath); err != nil {
			t.Fatalf("failed to remove output file before Render test: %v", err)
		}
	}

	r := Renderer{
		ImageInline: false,
		Template:    "<html>\n<body>\n{{{content}}}\n</body>\n</html>",
		BaseDir:     basePath,
		OutDir:      outDir,
	}

	if err := r.Render(mdPath); err != nil {
		t.Fatalf("Render unexpectedly gave an error: %v", err)
	}

	content, err := ioutil.ReadFile(outFilePath)

	if err != nil {
		t.Errorf("Render did not seem to write html file. Failed to read the output file after Renderer reports success.: %v", err)
	}

	if len(string(content)) == 0 {
		t.Error("Render did not seem to write html file. File content does not exists (0 byte).")
	}

	// Exported html content won't be validated here since a depended library
	// convert markdown content to html.
}

func TestOutPath(t *testing.T) {
	type TestCase struct {
		InputFile string
		OutDir    string
		BaseDir   string
		Expected  string
	}

	testCases := []TestCase{
		TestCase{
			"C:\\foo\\bar\\baz.md",
			"C:\\qux\\quux",
			"C:\\foo",
			"C:\\qux\\quux\\bar\\baz.html",
		},
		TestCase{
			"C:\\あいうえお\\bar\\baz.md",
			"C:\\qux\\quux",
			"C:\\あいうえお",
			"C:\\qux\\quux\\bar\\baz.html",
		},
		TestCase{
			"C:\\あいうえお\\かきくけこ\\foo\\さしすせそ.md",
			"C:\\たちつてと",
			"C:\\あいうえお",
			"C:\\たちつてと\\かきくけこ\\foo\\さしすせそ.html",
		},
	}

	for i, testCase := range testCases {
		got := outPath(testCase.InputFile, testCase.OutDir, testCase.BaseDir)
		if got != testCase.Expected {
			t.Errorf("\n%di\ngot %v\nwant %v", i, got, testCase.Expected)
		}
	}
}

func TestChangeExtension(t *testing.T) {
	got := changeExtension("C:\\foo\\bar.md", "html")
	want := "C:\\foo\\bar.html"
	if got != want {
		t.Errorf("\ngot %v\nwant %v", got, want)
	}
}

func TestDropExtension(t *testing.T) {
	got := dropExtension("C:\\foo\\bar.md")
	want := "C:\\foo\\bar"
	if got != want {
		t.Errorf("\ngot %v\nwant %v", got, want)
	}
}

func TestImageToBase64(t *testing.T) {
	curPath, _ := os.Getwd()
	assetsPath := filepath.Join(curPath, "..", "test_assets", "image")

	got, err := imageToBase64(filepath.Join(assetsPath, "company.png"))
	want, _ := ioutil.ReadFile(filepath.Join(assetsPath, "base64_company.txt"))

	if err != nil {
		t.Errorf("\nimageToBase64 unexpectedly gave an error: %v", err)
	}
	if got != string(want) {
		t.Errorf("\nbase64 encoding of image file went wrong.")
	}
}

func TestImageToBase64FailByFileNotFound(t *testing.T) {
	_, err := imageToBase64("C:\\a_file_that_should_not_exist")
	if err == nil {
		t.Errorf("imageToBase64 unexpectedly gave no error while input file does not exist.")
	}
}

func TestCopyFileNoNothingSinceSameDestination(t *testing.T) {
	err := copyFile("C:\\foo\\bar.png", "C:\\foo\\bar.png")
	if err != nil {
		t.Errorf("copyFile gave an error (It shouldn't when source and destination path are the same): %v", err)
	}
}

func TestCopyFileSrcNotFound(t *testing.T) {
	curPath, _ := os.Getwd()
	assetsPath := filepath.Join(curPath, "..", "test_assets", "image")

	src := filepath.Join(assetsPath, "company_not_exists.png")
	dst := filepath.Join(assetsPath, "company_copy.png")
	err := copyFile(src, dst)

	if err == nil {
		t.Errorf("copyFile gave no error even though input file does not exist: %v", src)
	}
}

func TestCopyFileDestDirNotExist(t *testing.T) {
	curPath, _ := os.Getwd()
	assetsPath := filepath.Join(curPath, "..", "test_assets", "image")

	src := filepath.Join(assetsPath, "company.png")
	dst := filepath.Join(assetsPath, "unknown_dir", "company_copy.png")
	err := copyFile(src, dst)

	if err == nil {
		t.Errorf("copyFile gave no error even though destination directory does not exist: %v", dst)
	}
}

func TestCopyFile(t *testing.T) {
	curPath, _ := os.Getwd()
	assetsPath := filepath.Join(curPath, "..", "test_assets", "image")

	src := filepath.Join(assetsPath, "company.png")
	dst := filepath.Join(assetsPath, "company_copy.png")

	// cleaning up the destination file
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			t.Fatalf("failed to remove destination file before copy: %v", err)
		}
	}

	err := copyFile(src, dst)

	if err != nil {
		t.Errorf("copyFile gave an error: %v", dst)
	}

	_, statErr := os.Stat(dst)
	if statErr != nil {
		t.Errorf("copyFile did not seem to copy file: %v", statErr)
	}
}
