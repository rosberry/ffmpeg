package ffmpeg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type (
	Settings struct {
		FFMPEGPath        string `json:"ffmpegPath"`
		SrcPath           string `json:"srcPath"`
		TempPath          string `json:"tempPath"`
		SkipStandardTests bool
		FileSettings      []FileConfig
	}
	FileConfig struct {
		StandardType    bool
		Filename        string
		Title           string
		Duration        uint
		TrimDuration    int
		TrimStart       int
		Bitrate         uint
		Width           uint
		Height          uint
		ThumbnailWidth  int
		ThumbnailHeight int
	}
)

const (
	expectedDurationTime = 10
	expectedDurationUnit = time.Second
	testVideoFileName    = "testVideo.avi"
	testBitrateFileName  = "testBitrate.avi"
	testVideoTitle       = "test title"
	testBitrate          = 4000
	testWidth            = 1280
	testHeight           = 720
	defaultStartTrim     = 2
	durationTrim         = 5
)

var (
	checkSettingsFile bool
	hasTempPath       bool
	settings          = Settings{
		FileSettings: []FileConfig{},
	}

	errorJSONParse       = errors.New("error json parse")
	errorRegularSubmatch = errors.New("error regular submatch")
	errorUndefinedPath   = errors.New("path is undefined")

	pathToConfigJSON = os.Getenv("JSON_CONFIG_PATH")
	pathToUtility    = os.Getenv("FFMPEG_SRC")
	pathToSources    = os.Getenv("TEST_FILES_SRC")

	expectedDuration = expectedDurationTime * expectedDurationUnit
	conveyTitles     = map[string]map[string]string{
		"New": {
			"given": "Given ffmpeg path.",
			"when":  "When we call the function 'new'.",
			"then":  "Then we expect the correct return type.",
		},
		"SetPath": {
			"given": "Given test path to video sources and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'SetPath'.",
			"then":  "Then path to the resource folder is valid \nAnd path to source should be equal path in ffmpeg settings \nAnd we do not have any errors related to the resource folder.",
		},
		"Duration": {
			"given": "Given test path to test video '%s' and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Duration'.",
			"then":  "Then we have no errors with receiving data \nAnd the duration in video should be equal duration in response.",
		},
		"SimpleDuration": {
			"given": "Given test path to test video '%s' and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'SimpleDuration'.",
			"then":  "Then we have no errors with receiving data \nAnd the duration in video should be equal duration in response.",
		},
		"Trim": {
			"given": "Given test path to test video '%s' and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Trim'.",
			"then":  "Then response don't should return error \nAnd running function 'Duration' for check value don't should return error\nAnd duration in response should be equal expected duration value.",
		},
		"Bitrate": {
			"given": "Given test path to test video '%s' by bitrate and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Bitrate'.",
			"then":  "Then response don't should return error \nAnd bitrate of the result is equal to the test value up to a hundred.",
		},
		"Title": {
			"given": "Given test path to test video '%s' and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Title'.",
			"then":  "Then response don't should return error.",
		},
		"Size": {
			"given": "Given test path to test video '%s' and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Size'.",
			"then":  "Then response don't should return error \nAnd width in response should be equal width in test value \nAnd height in response should be equal height in test value.",
		},
		"Thumbnail": {
			"given": "Given test path to test video '%s' and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Thumbnail'.",
			"then":  "Then response don't should return error.",
		},
	}
)

func setConfig() error {
	if pathToConfigJSON != "" {
		checkSettingsFile = true
		jsonFile, err := os.Open(pathToConfigJSON)
		defer jsonFile.Close()

		if err != nil {
			return err
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)
		err = json.Unmarshal(byteValue, &settings)

		if err != nil {
			return fmt.Errorf("%w: %s", errorJSONParse, err)
		}

		if err := setBaseSrc(settings.FFMPEGPath, settings.SrcPath, settings.TempPath); err != nil {
			return err
		}

		if !settings.SkipStandardTests {
			defaultConfig := getDefaultConfig()
			settings.FileSettings = append(defaultConfig, settings.FileSettings...)
		}
	} else {
		if err := setBaseSrc(pathToUtility, pathToSources, pathToSources); err != nil {
			return err
		}

		settings.FileSettings = getDefaultConfig()
	}
	return nil
}

func setBaseSrc(ffmpegSrc string, srcPath string, tempPath string) error {
	settings.FFMPEGPath = ffmpegSrc
	if settings.FFMPEGPath == "" {
		settings.FFMPEGPath = "ffmpeg"
	}

	if err := ffmpegExists(settings.FFMPEGPath); err != nil {
		return err
	}

	settings.SrcPath = srcPath
	if settings.SrcPath == "" {
		settings.SrcPath = "./"
	}

	if err := pathExists(settings.SrcPath); err != nil {
		return err
	}

	if !settings.SkipStandardTests {
		settings.TempPath = tempPath
		if settings.TempPath == "" {
			dir, err := ioutil.TempDir("", "testPath")
			if err != nil {
				return err
			}
			settings.TempPath = dir + "/"
			hasTempPath = true
		}
	}

	return nil
}

func getDefaultConfig() []FileConfig {
	config := []FileConfig{
		{
			StandardType:    true,
			Filename:        testVideoFileName,
			Title:           testVideoTitle,
			Duration:        expectedDurationTime,
			TrimDuration:    durationTrim,
			TrimStart:       defaultStartTrim,
			Width:           testWidth,
			Height:          testHeight,
			ThumbnailWidth:  testWidth,
			ThumbnailHeight: testHeight,
		},
		{
			StandardType: true,
			Filename:     testBitrateFileName,
			Bitrate:      testBitrate,
		},
	}

	return config
}

func TestMain(m *testing.M) {
	if err := setConfig(); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	if !checkSettingsFile || (checkSettingsFile && !settings.SkipStandardTests) {
		createBaseVideo()
		cloneVideoForBitrate()

		defer removeFile(settings.SrcPath + testVideoFileName)
		defer removeFile(settings.SrcPath + testBitrateFileName)
	}

	if hasTempPath && settings.TempPath != "" {
		defer os.RemoveAll(settings.TempPath)
	}

	m.Run()
}

func TestNew(test *testing.T) {
	Convey(conveyTitles["New"]["given"], test, func() {
		Convey(conveyTitles["New"]["when"], func() {
			ffmpeg := New()
			Convey(conveyTitles["New"]["then"], func() {
				resType := fmt.Sprintf("%T", ffmpeg)
				checkType := "*ffmpeg.FFMpeg"
				So(resType, ShouldEqual, checkType)
			})
		})
	})
}

func TestSetPath(test *testing.T) {
	Convey(conveyTitles["SetPath"]["given"], test, func() {
		Convey(conveyTitles["SetPath"]["when"], func() {
			ffmpeg := New().SetPath(settings.FFMPEGPath)
			Convey(conveyTitles["SetPath"]["then"], func() {
				So(settings.FFMPEGPath, ShouldEqual, ffmpeg.path)
			})
		})
	})
}

func TestDuration(test *testing.T) {
	for _, config := range settings.FileSettings {
		if config.Duration != 0 {
			given := fmt.Sprintf(conveyTitles["Duration"]["given"], config.Filename)
			Convey(given, test, func() {
				ffmpeg := initFFMPEG()
				filePath := getFilePath(config)

				Convey(conveyTitles["Duration"]["when"], func() {
					duration, err := ffmpeg.Duration(filePath)

					Convey(conveyTitles["Duration"]["then"], func() {
						So(err, ShouldEqual, nil)
						if err == nil {
							actualDuration := duration.String()
							expectedDuration := (expectedDurationUnit * time.Duration(config.Duration)).String()
							So(actualDuration, ShouldEqual, expectedDuration)
						}
					})
				})
			})
		}
	}
}

func TestSimpleDuration(test *testing.T) {
	for _, config := range settings.FileSettings {
		if config.Duration != 0 {
			given := fmt.Sprintf(conveyTitles["SimpleDuration"]["given"], config.Filename)
			Convey(given, test, func() {
				ffmpeg := initFFMPEG()
				filePath := getFilePath(config)

				Convey(conveyTitles["SimpleDuration"]["when"], func() {
					duration, err := ffmpeg.SimpleDuration(filePath)

					Convey(conveyTitles["SimpleDuration"]["then"], func() {
						So(err, ShouldEqual, nil)
						if err == nil {
							actualDuration := duration.String()
							expectedDuration := (expectedDurationUnit * time.Duration(config.Duration)).String()
							So(actualDuration, ShouldEqual, expectedDuration)
						}
					})
				})
			})
		}
	}
}

func TestTrim(test *testing.T) {
	for _, config := range settings.FileSettings {
		if config.TrimDuration != 0 {
			given := fmt.Sprintf(conveyTitles["Trim"]["given"], config.Filename)
			Convey(given, test, func() {
				ffmpeg := initFFMPEG()
				filePath := getFilePath(config)

				Convey(conveyTitles["Trim"]["when"], func() {
					trimFilePath := getBasePath(config.StandardType) + "testVideoTrim.mpg"

					startTrim := config.TrimStart
					if startTrim == 0 {
						startTrim = defaultStartTrim
					}

					err := ffmpeg.TrimVideo(filePath, trimFilePath, startTrim, config.TrimDuration)
					defer removeFile(trimFilePath)

					Convey(conveyTitles["Trim"]["then"], func() {
						So(err, ShouldEqual, nil)

						duration, err := ffmpeg.Duration(trimFilePath)
						So(err, ShouldEqual, nil)

						actualDuration := duration.Round(time.Second).String()
						expectedDuration := (time.Second * time.Duration(config.TrimDuration)).String()
						So(actualDuration, ShouldEqual, expectedDuration)
					})
				})
			})
		}
	}
}

func TestBitrate(test *testing.T) {
	for _, config := range settings.FileSettings {
		if config.Bitrate != 0 {
			given := fmt.Sprintf(conveyTitles["Bitrate"]["given"], config.Filename)
			Convey(given, test, func() {
				ffmpeg := initFFMPEG()
				filePath := getFilePath(config)

				Convey(conveyTitles["Bitrate"]["when"], func() {
					bitrate, err := ffmpeg.Bitrate(filePath)
					// todo regular

					Convey(conveyTitles["Bitrate"]["then"], func() {
						So(err, ShouldEqual, nil)
						if err != nil {
							return
						}

						reg := regexp.MustCompile(`(\d+) [/\w]+`)
						res := reg.FindStringSubmatch(*bitrate)
						if len(res) < 2 {
							So(errorRegularSubmatch, ShouldEqual, nil)
							return
						}
						bitrateStr := res[1]
						bitrateInt, _ := strconv.Atoi(bitrateStr)
						bitrateInt = int(math.Round(float64(bitrateInt)/100) * 100)

						So(bitrateInt, ShouldEqual, config.Bitrate)
					})
				})
			})
		}
	}
}

func TestTitle(test *testing.T) {
	for _, config := range settings.FileSettings {
		if config.Title != "" {
			given := fmt.Sprintf(conveyTitles["Title"]["given"], config.Filename)
			Convey(given, test, func() {
				ffmpeg := initFFMPEG()
				filePath := getFilePath(config)

				Convey(conveyTitles["Title"]["when"], func() {
					title, err := ffmpeg.Title(filePath)

					Convey(conveyTitles["Title"]["then"], func() {
						So(err, ShouldEqual, nil)
						So(*title, ShouldEqual, config.Title)
					})
				})
			})
		}
	}
}

func TestSize(test *testing.T) {
	for _, config := range settings.FileSettings {
		if config.Width != 0 && config.Height != 0 {
			given := fmt.Sprintf(conveyTitles["Size"]["given"], config.Filename)
			Convey(given, test, func() {
				ffmpeg := initFFMPEG()
				filePath := getFilePath(config)

				Convey(conveyTitles["Size"]["when"], func() {
					width, height, err := ffmpeg.Size(filePath)

					Convey(conveyTitles["Size"]["then"], func() {
						So(err, ShouldEqual, nil)
						So(width, ShouldEqual, config.Width)
						So(height, ShouldEqual, config.Height)
					})
				})
			})
		}
	}
}

func TestThumbnail(test *testing.T) {
	for _, config := range settings.FileSettings {
		if config.ThumbnailWidth != 0 && config.ThumbnailHeight != 0 {
			given := fmt.Sprintf(conveyTitles["Thumbnail"]["given"], config.Filename)
			Convey(given, test, func() {
				ffmpeg := initFFMPEG()
				filePath := getFilePath(config)

				Convey(conveyTitles["Thumbnail"]["when"], func() {
					thumbnailFilePath := getBasePath(config.StandardType) + "testThumbnail.jpeg"

					err := ffmpeg.CreateThumbnail(filePath, thumbnailFilePath, config.ThumbnailWidth, config.ThumbnailHeight)
					defer removeFile(thumbnailFilePath)

					Convey(conveyTitles["Thumbnail"]["then"], func() {
						So(err, ShouldEqual, nil)
					})
				})
			})
		}
	}
}

func ffmpegExists(path string) error {
	cmd := exec.Command(path)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	out := &stdout
	if err := cmd.Run(); err != nil {
		out = &stderr
	}

	result := regexp.MustCompile(`ffmpeg version`).FindSubmatch(out.Bytes())
	if len(result) < 1 {
		return fmt.Errorf("%w: %s", errorUndefinedPath, path)
	}
	return nil
}

func pathExists(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", errorUndefinedPath, path)
	}
	return nil
}

func createBaseVideo() error {
	srcConf := fmt.Sprintf("testsrc=duration=%d:size=%dx%d", expectedDurationTime, testWidth, testHeight)
	title := fmt.Sprintf("title=%s", testVideoTitle)
	basePath := settings.TempPath + testVideoFileName

	cmd := exec.Command(settings.FFMPEGPath, "-f", "lavfi", "-i", srcConf, "-metadata", title, "-y", basePath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Start()

	err := cmd.Wait()
	if err != nil {
		return fmt.Errorf("%s :%w", "Error creating file", err)
	}

	output := stderr.String()
	log.Print(output)
	return nil
}

func cloneVideoForBitrate() error {
	bitrateStr := fmt.Sprintf("%dk", testBitrate)
	basePath := settings.TempPath + testVideoFileName
	outputPath := settings.TempPath + testBitrateFileName

	cmd := exec.Command(settings.FFMPEGPath, "-i", basePath, "-y", "-b", bitrateStr, "-minrate", bitrateStr, "-maxrate", bitrateStr, outputPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Start()

	err := cmd.Wait()
	if err != nil {
		return fmt.Errorf("%s :%w", "Error creating file", err)
	}

	output := stderr.String()
	log.Print(output)
	return nil
}

func initFFMPEG() *FFMpeg {
	ffmpeg := New().SetPath(settings.FFMPEGPath)
	return ffmpeg
}

func removeFile(path string) {
	_, err := os.Stat(path)
	if err == nil {
		os.Remove(path)
	}
}

func getBasePath(isStandardType bool) string {
	if isStandardType {
		return settings.TempPath
	} else {
		return settings.SrcPath
	}
}

func getFilePath(config FileConfig) string {
	return getBasePath(config.StandardType) + config.Filename
}
