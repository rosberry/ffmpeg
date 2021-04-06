package ffmpeg

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	expectedDurationTime = 10
	expectedDurationUnit = time.Second
	testVideoFilePath    = "/testVideo.avi"
	testBitrateFilePath  = "/testBitrate.avi"
	testVideoTitle       = "test title"
	testBitrate          = 4000
	testWidth            = 1280
	testHeight           = 720
	startTrim            = 2
	durationTrim         = 5
)

var (
	errorUndefinedPath = errors.New("path is undefined")

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
			"given": "Given test path to test video and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Duration'.",
			"then":  "Then we have no errors with receiving data \nAnd the duration in video should be equal duration in response.",
		},
		"SimpleDuration": {
			"given": "Given test path to test video and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'SimpleDuration'.",
			"then":  "Then we have no errors with receiving data \nAnd the duration in video should be equal duration in response.",
		},
		"Trim": {
			"given": "Given test path to test video and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Trim'.",
			"then":  "Then response don't should return error \nAnd running function 'Duration' for check value don't should return error\nAnd duration in response should be equal expected duration value.",
		},
		"Bitrate": {
			"given": "Given test path to test video by bitrate and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Bitrate'.",
			"then":  "Then response don't should return error \nAnd bitrate of the result is equal to the test value up to a hundred.",
		},
		"Title": {
			"given": "Given test path to test video and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Title'.",
			"then":  "Then response don't should return error.",
		},
		"Size": {
			"given": "Given test path to test video and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Size'.",
			"then":  "Then response don't should return error \nAnd width in response should be equal width in test value \nAnd height in response should be equal height in test value.",
		},
		"Thumbnail": {
			"given": "Given test path to test video and ffmpeg path.",
			"when":  "When we create an ffmpeg instance and run a function 'Thumbnail'.",
			"then":  "Then response don't should return error.",
		},
	}
)

func TestMain(m *testing.M) {
	if err := ffmpegExists(pathToUtility); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	if err := pathExists(pathToSources); err != nil {
		log.Print(err)
		os.Exit(1)
	}

	createBaseVideo()
	cloneVideoForBitrate()

	defer removeFile(pathToSources + testVideoFilePath)
	defer removeFile(pathToSources + testBitrateFilePath)

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
			ffmpeg := New().SetPath(pathToUtility)
			Convey(conveyTitles["SetPath"]["then"], func() {
				So(pathToUtility, ShouldEqual, ffmpeg.path)
			})
		})
	})
}

func TestDuration(test *testing.T) {
	Convey(conveyTitles["Duration"]["given"], test, func() {
		Convey(conveyTitles["Duration"]["when"], func() {
			ffmpeg := initFFMPEG()
			duration, err := ffmpeg.Duration(pathToSources + testVideoFilePath)
			Convey(conveyTitles["Duration"]["then"], func() {
				So(err, ShouldEqual, nil)
				if err == nil {
					actualDuration := duration.String()
					expectedDuration := expectedDuration.String()
					So(actualDuration, ShouldEqual, expectedDuration)
				}
			})
		})
	})
}

func TestSimpleDuration(test *testing.T) {
	Convey(conveyTitles["SimpleDuration"]["given"], test, func() {
		Convey(conveyTitles["SimpleDuration"]["when"], func() {
			ffmpeg := initFFMPEG()
			sDuration, err := ffmpeg.SimpleDuration(pathToSources + testVideoFilePath)
			Convey(conveyTitles["SimpleDuration"]["then"], func() {
				So(err, ShouldEqual, nil)
				if err == nil {
					actualDuration := sDuration.String()
					expectedDuration := expectedDuration.String()
					So(actualDuration, ShouldEqual, expectedDuration)
				}
			})
		})
	})
}

func TestTrim(test *testing.T) {
	Convey(conveyTitles["Trim"]["given"], test, func() {
		Convey(conveyTitles["Trim"]["when"], func() {
			ffmpeg := initFFMPEG()
			testPath := pathToSources + testVideoFilePath
			trimFilePath := pathToSources + "/testVideoTrim.mpg"

			err := ffmpeg.TrimVideo(testPath, trimFilePath, startTrim, durationTrim)
			defer removeFile(trimFilePath)

			Convey(conveyTitles["Trim"]["then"], func() {
				So(err, ShouldEqual, nil)

				duration, err := ffmpeg.Duration(trimFilePath)
				So(err, ShouldEqual, nil)

				actualDuration := duration.Round(time.Second).String()
				expectedDuration := (time.Second * durationTrim).String()
				So(actualDuration, ShouldEqual, expectedDuration)
			})
		})
	})
}

func TestBitrate(test *testing.T) {
	Convey(conveyTitles["Bitrate"]["given"], test, func() {
		Convey(conveyTitles["Bitrate"]["when"], func() {
			ffmpeg := initFFMPEG()
			testPath := pathToSources + testBitrateFilePath
			bitrate, err := ffmpeg.Bitrate(testPath)

			bitrateStr := strings.Replace(*bitrate, " kb/s", "", 1)
			bitrateInt, _ := strconv.Atoi(bitrateStr)

			Convey(conveyTitles["Bitrate"]["then"], func() {
				So(err, ShouldEqual, nil)

				bitrateInt = int(math.Round(float64(bitrateInt)/100) * 100)
				So(bitrateInt, ShouldEqual, testBitrate)
			})
		})
	})
}

func TestTitle(test *testing.T) {
	Convey(conveyTitles["Title"]["given"], test, func() {
		Convey(conveyTitles["Title"]["when"], func() {
			ffmpeg := initFFMPEG()
			testPath := pathToSources + testVideoFilePath
			title, err := ffmpeg.Title(testPath)

			Convey(conveyTitles["Title"]["then"], func() {
				So(err, ShouldEqual, nil)
				So(*title, ShouldEqual, testVideoTitle)
			})
		})
	})
}

func TestSize(test *testing.T) {
	Convey(conveyTitles["Size"]["given"], test, func() {
		Convey(conveyTitles["Size"]["when"], func() {
			ffmpeg := initFFMPEG()
			testPath := pathToSources + testVideoFilePath
			width, height, err := ffmpeg.Size(testPath)

			Convey(conveyTitles["Size"]["then"], func() {
				So(err, ShouldEqual, nil)
				So(width, ShouldEqual, testWidth)
				So(height, ShouldEqual, testHeight)
			})
		})
	})
}

func TestThumbnail(test *testing.T) {
	Convey(conveyTitles["Thumbnail"]["given"], test, func() {
		Convey(conveyTitles["Thumbnail"]["when"], func() {
			ffmpeg := initFFMPEG()
			testPath := pathToSources + testVideoFilePath
			thumbnailFilePath := pathToSources + "/testThumbnail.jpeg"

			err := ffmpeg.CreateThumbnail(testPath, thumbnailFilePath, testWidth, testHeight)
			defer removeFile(thumbnailFilePath)

			Convey(conveyTitles["Thumbnail"]["then"], func() {
				So(err, ShouldEqual, nil)
			})
		})
	})
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
	cmd := exec.Command(pathToUtility, "-f", "lavfi", "-i", srcConf, "-metadata", title, "-y", pathToSources+testVideoFilePath)
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
	cmd := exec.Command(pathToUtility, "-i", pathToSources+testVideoFilePath, "-y", "-b", bitrateStr, "-minrate", bitrateStr, "-maxrate", bitrateStr, pathToSources+testBitrateFilePath)
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
	ffmpeg := New().SetPath(pathToUtility)
	return ffmpeg
}

func removeFile(path string) {
	_, err := os.Stat(path)
	if err == nil {
		os.Remove(path)
	}
}
