package ffmpeg

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	TestExpectedDuration = time.Second * 10
	TestDurationFilename = "/testDuration.mpg"
)

func TestStart(test *testing.T) {
	Convey("Given test path to video sources and path to ffmpeg binary", test, func(convey C) {
		pathToUtility := os.Getenv("FFMPEG_SRC")
		pathToSources := os.Getenv("TEST_FILES_SRC")
		var abortProcess bool
		convey.Convey("We want to check that the utility correctly creates a new instance", func(convey C) {
			ffmpeg := New()
			resType := fmt.Sprintf("%T", ffmpeg)
			convey.Convey("We want to check that the path to the return type matches the expectation", func() {
				checkType := "*ffmpeg.FFMpeg"
				if resType != checkType {
					abortProcess = true
				}
				convey.So(resType, ShouldEqual, checkType)
			})
		})
		if abortProcess {
			return
		}

		convey.Convey("We want to check that the utility correctly set path of video sources", func(convey C) {
			ffmpeg := New()
			res, err := ffmpeg.SetPath(pathToSources)

			convey.Convey("We want to check that the path to the resource folder we have set is valid", func() {
				if err != nil {
					abortProcess = true
				}
				convey.So(err, ShouldEqual, nil)
			})
			if err != nil {
				return
			}

			convey.Convey("We want to check that path to source should be Equal path in ffmpeg settings", func() {
				convey.So(pathToSources, ShouldEqual, res.path)
			})
			if pathToSources != res.path {
				abortProcess = true
			}

			convey.Convey("We want to check that we do not have any errors related to the resource folder", func() {
				_, err := os.Stat(pathToSources)
				convey.So(err, ShouldEqual, nil)
				convey.So(os.IsNotExist(err), ShouldEqual, false)
			})
		})
		if abortProcess {
			return
		}

		convey.Convey("We want to check that the duration parameter is returned correctly.", func(convey C) {
			ffmpeg := New()
			_, _ = ffmpeg.SetPath(pathToUtility)
			duration, err := ffmpeg.SimpleDuration(pathToSources + TestDurationFilename)

			convey.Convey("We want to check that the duration in video should be equal duration in response and we have no errors with receiving data", func() {
				convey.So(err, ShouldEqual, nil)
				if err != nil {
					return
				}
				actualDuration := duration.String()
				ExpectedDuration := time.Duration(TestExpectedDuration).String()
				convey.So(actualDuration, ShouldEqual, ExpectedDuration)
			})
		})
	})
}
