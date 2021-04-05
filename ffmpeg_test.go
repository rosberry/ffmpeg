package ffmpeg

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	TestExpectedDuration = time.Second * 10
	TestVideoFilePath    = "/testVideo.mpg"
	TestBitrateFilePath  = "/testBitrate.mpg"
	TestTitleFilePath    = "/testTitle.mpg"
	TestBitrate          = 4000
	TestWidth            = 1280
	TestHeight           = 720
	startTrim            = 2
	durationTrim         = 5
)

var (
	pathToUtility = os.Getenv("FFMPEG_SRC")
	pathToSources = os.Getenv("TEST_FILES_SRC")
	abortProcess  bool
	conveyGiven   = "Given test path to video sources and path to ffmpeg binary"
	conveyTitles  = map[string]map[string]string{
		"new": {
			"title":      "We want to check that the utility correctly creates a new instance.",
			"check type": "We want to check that the path to the return type matches the expectation.",
		},
		"setPath": {
			"title":        "We want to check that the utility correctly set path of video sources.",
			"valid path":   "We want to check that the path to the resource folder we have set is valid.",
			"equal path":   "We want to check that path to source should be equal path in ffmpeg settings.",
			"other errors": "We want to check that we do not have any errors related to the resource folder.",
		},
		"duration": {
			"title":        "We want to check that the 'Duration' and 'SimpleDuration' functions is returned correctly values.",
			"check simple": "We want to check that the duration in video should be equal duration in response 'SimpleDuration' function and we have no errors with receiving data.",
			"check full":   "We want to check that the duration in video should be equal duration in response 'Duration' function and we have no errors with receiving data.",
		},
		"trim": {
			"title":          "We want to check that trim video function is work correctly.",
			"no error":       "We want to check that function do not return error.",
			"check duration": "We want to check that created trim video in correct duration.",
		},
		"bitrate": {
			"title":    "We want to check that the 'Bitrate' function is work correctly.",
			"no error": "We want to check that function do not return error.",
			"equal":    "We want to check that result bitrate be equal test value",
		},
		"title": {
			"title":    "We want to check that the 'Title' function is work correctly.",
			"no error": "We want to check that function do not return error.",
		},
		"size": {
			"title": "We want to check that the 'Size' function is work correctly.",
			"equal": "We want to check that the width and height parameters should be equal test sizes",
		},
		"thumbnail": {
			"title":    "We want to check that the 'CreateThumbnail' function is work correctly.",
			"no error": "We want to check that function do not return error.",
		},
	}
)

// func TestMain(main *testing)
func TestStart(test *testing.T) {
	// createTestSrc()
	Convey(conveyGiven, test, func() {
		Convey(conveyTitles["new"]["title"], func() {
			ffmpeg := New()
			resType := fmt.Sprintf("%T", ffmpeg)
			Convey(conveyTitles["new"]["check type"], func() {
				checkType := "*ffmpeg.FFMpeg"
				if resType != checkType {
					abortProcess = true
				}
				So(resType, ShouldEqual, checkType)
			})
		})
		if abortProcess {
			return
		}

		Convey(conveyTitles["setPath"]["title"], func() {
			ffmpeg, err := initFFMPEG()

			Convey(conveyTitles["setPath"]["valid path"], func() {
				if err != nil {
					abortProcess = true
				}
				So(err, ShouldEqual, nil)
			})
			if err != nil {
				return
			}

			Convey(conveyTitles["setPath"]["equal path"], func() {
				So(pathToUtility, ShouldEqual, ffmpeg.path)
			})
			if pathToUtility != ffmpeg.path {
				abortProcess = true
			}

			Convey(conveyTitles["setPath"]["other errors"], func() {
				_, err := os.Stat(pathToUtility)
				So(err, ShouldEqual, nil)
				So(os.IsNotExist(err), ShouldEqual, false)
			})
		})
		if abortProcess {
			return
		}

		Convey(conveyTitles["duration"]["title"], func() {
			ffmpeg, _ := initFFMPEG()
			sDuration, err := ffmpeg.SimpleDuration(pathToSources + TestVideoFilePath)

			Convey(conveyTitles["duration"]["check simple"], func() {
				So(err, ShouldEqual, nil)
				if err != nil {
					return
				}
				actualDuration := sDuration.String()
				expectedDuration := TestExpectedDuration.String()
				So(actualDuration, ShouldEqual, expectedDuration)
			})

			duration, err := ffmpeg.Duration(pathToSources + TestVideoFilePath)

			Convey(conveyTitles["duration"]["check full"], func() {
				So(err, ShouldEqual, nil)
				if err != nil {
					return
				}
				actualDuration := duration.String()
				expectedDuration := TestExpectedDuration.String()
				So(actualDuration, ShouldEqual, expectedDuration)
			})
		})

		Convey(conveyTitles["trim"]["title"], func() {
			ffmpeg, _ := initFFMPEG()
			testPath := pathToSources + TestVideoFilePath
			trimFilePath := pathToSources + "/testVideoTrim.mpg"

			err := ffmpeg.TrimVideo(testPath, trimFilePath, startTrim, durationTrim)
			defer removeFile(trimFilePath)

			Convey(conveyTitles["trim"]["no error"], func() {
				So(err, ShouldEqual, nil)
			})

			Convey(conveyTitles["trim"]["check duration"], func() {
				duration, err := ffmpeg.Duration(trimFilePath)
				So(err, ShouldEqual, nil)
				actualDuration := duration.String()
				expectedDuration := (time.Second * durationTrim).String()
				So(actualDuration, ShouldEqual, expectedDuration)
			})
		})

		Convey(conveyTitles["bitrate"]["title"], func() {
			ffmpeg, _ := initFFMPEG()
			testPath := pathToSources + TestBitrateFilePath
			bitrate, err := ffmpeg.Bitrate(testPath)

			bitrateStr := strings.Replace(*bitrate, " kb/s", "", 1)
			bitrateInt, _ := strconv.Atoi(bitrateStr)

			Convey(conveyTitles["bitrate"]["no error"], func() {
				So(err, ShouldEqual, nil)
			})

			Convey(conveyTitles["bitrate"]["equal"], func() {
				So(bitrateInt, ShouldEqual, TestBitrate)
			})
		})
		// convey.Convey(conveyTitles["title"]["title"], func(convey C) {
		// 	ffmpeg, _ := initFFMPEG()
		// 	testPath := pathToSources + TestTitleFilePath
		// 	title, err := ffmpeg.Title(testPath)
		// 	log.Print(title)

		// 	convey.Convey(conveyTitles["title"]["no error"], func() {
		// 		convey.So(err, ShouldEqual, nil)
		// 	})
		// })

		Convey(conveyTitles["size"]["title"], func() {
			ffmpeg, _ := initFFMPEG()
			testPath := pathToSources + TestVideoFilePath
			width, height, err := ffmpeg.Size(testPath)

			Convey(conveyTitles["size"]["no error"], func() {
				So(err, ShouldEqual, nil)
			})

			Convey(conveyTitles["size"]["equal"], func() {
				So(width, ShouldEqual, TestWidth)
				So(height, ShouldEqual, TestHeight)
			})
		})

		Convey(conveyTitles["thumbnail"]["title"], func() {
			ffmpeg, _ := initFFMPEG()
			testPath := pathToSources + TestVideoFilePath
			thumbnailFilePath := pathToSources + "/testThumbnail.jpeg"

			err := ffmpeg.CreateThumbnail(testPath, thumbnailFilePath, TestWidth, TestHeight)
			defer removeFile(thumbnailFilePath)

			Convey(conveyTitles["thumbnail"]["no error"], func() {
				So(err, ShouldEqual, nil)
			})
		})
	})
}

func initFFMPEG() (*FFMpeg, error) {
	ffmpeg := New()
	res, err := ffmpeg.SetPath(pathToUtility)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func removeFile(path string) {
	_, err := os.Stat(path)
	if err == nil {
		os.Remove(path)
	}
}
