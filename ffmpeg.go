package ffmpeg

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

type (
	FFMpeg struct {
		path string
	}
)

var (
	//ErrUnknownDuration is error "Could not determine the duration"
	ErrUnknownDuration = errors.New("Could not determine the duration")
	//ErrUnknownBitrate is error "Could not determine the bitrate"
	ErrUnknownBitrate = errors.New("Could not determine the bitrate")
	//ErrUnknownTitle is error "Could not determine the title"
	ErrUnknownTitle = errors.New("Could not determine the title")
	//ErrUnknownSize is error "Could not determine the size"
	ErrUnknownSize = errors.New("Could not determine the size")
	//ErrCannotCreateThumbnail is error "Could not create thumbnail"
	ErrCannotCreateThumbnail = errors.New("Could not create thumbnail")
	//ErrCannotTrimVideo is error "Could not trim video"
	ErrCannotTrimVideo = errors.New("Could not trim video")
)

//New creates a ffmpeg object with default path ("../other/ffmpeg")
func New() *FFMpeg {
	return &FFMpeg{
		path: "../other/ffmpeg",
	}
}

//SetPath set custom path to ffmpeg execution file
func (f *FFMpeg) SetPath(path string) *FFMpeg {
	f.path = path
	return f
}

//Duration get duration with encoding
func (f *FFMpeg) Duration(filePath string) (*time.Duration, error) {
	cmd := exec.Command(f.path, "-i", filePath, "-f", "null", "-")

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Start()
	cmd.Wait()
	output := string(stderr.Bytes())

	re := regexp.MustCompile(`time=([0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]+)`)
	var h, m int64
	var s float64

	results := re.FindAllString(output, -1)
	if len(results) < 1 {
		return nil, ErrUnknownDuration
	}
	result := results[len(results)-1]

	if _, err := fmt.Sscanf(result, `time=%d:%d:%f`, &h, &m, &s); err != nil {
		return nil, ErrUnknownDuration
	}
	d := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s*1000000000)
	return &d, nil
}

//SimpleDuration get duration without encoding (not for all codecs)
func (f *FFMpeg) SimpleDuration(filePath string) (*time.Duration, error) {
	cmd := exec.Command(f.path, "-i", filePath)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Start()
	cmd.Wait()
	output := string(stderr.Bytes())

	re := regexp.MustCompile(`Duration: ([0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]+)`)
	var h, m int64
	var s float64

	results := re.FindAllString(output, -1)
	if len(results) < 1 {
		// use thicc method as fallback
		return f.Duration(filePath)
		// return nil, ErrUnknownDuration
	}
	result := results[len(results)-1]

	if _, err := fmt.Sscanf(result, `Duration: %d:%d:%f`, &h, &m, &s); err != nil {
		// use thicc method as fallback
		return f.Duration(filePath)
		// return nil, ErrUnknownDuration
	}
	d := time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s*1000000000)
	return &d, nil
}

//TrimVideo trim video
func (f *FFMpeg) TrimVideo(filePath string, toFilePath string, start int, duration int) error {
	cmd := exec.Command(f.path, "-y", "-ss", fmt.Sprintf("%ds", start), "-i", filePath, "-to", fmt.Sprintf("%ds", duration), "-c", "copy", toFilePath)
	cmd.Start()
	if err := cmd.Wait(); err != nil {
		return ErrCannotTrimVideo
	}
	return nil
}

//Bitrate get file bitrate
func (f *FFMpeg) Bitrate(filePath string) (*string, error) {
	cmd := exec.Command(f.path, "-i", filePath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	out := &stdout
	if err := cmd.Run(); err != nil {
		out = &stderr
	}

	result := regexp.MustCompile(`bitrate: (.*)`).FindSubmatch(out.Bytes())
	if len(result) < 2 {
		return nil, ErrUnknownBitrate
	}
	s := string(result[1])
	return &s, nil
}

//Title get file title
func (f *FFMpeg) Title(filePath string) (*string, error) {
	cmd := exec.Command(f.path, "-i", filePath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	out := &stdout
	if err := cmd.Run(); err != nil {
		out = &stderr
	}

	result := regexp.MustCompile(`title\s*: (.*)`).FindSubmatch(out.Bytes())
	if len(result) < 2 {
		return nil, ErrUnknownTitle
	}
	s := string(result[1])
	return &s, nil
}

//Size get file size
func (f *FFMpeg) Size(filePath string) (uint, uint, error) {
	cmd := exec.Command(f.path, "-i", filePath)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	out := &stdout
	if err := cmd.Run(); err != nil {
		out = &stderr
	}

	result := regexp.MustCompile(`Stream.*Video.* ([0-9]+)x([0-9]+)[ ,]`).FindSubmatch(out.Bytes())
	if len(result) < 3 {
		return 0, 0, ErrUnknownSize
	}
	w, err := strconv.ParseUint(string(result[1]), 10, 64)
	if err != nil {
		return 0, 0, ErrUnknownSize
	}
	h, err := strconv.ParseUint(string(result[2]), 10, 64)
	if err != nil {
		return 0, 0, ErrUnknownSize
	}

	result = regexp.MustCompile(`displaymatrix: rotation of (-?\d*\.?\d*) degrees`).FindSubmatch(out.Bytes())
	if len(result) == 2 {
		if rotation, err := strconv.ParseFloat(string(result[1]), 64); err == nil {
			if int64(rotation)%180 != 0 {
				h, w = w, h
			}
		}
	}
	return uint(w), uint(h), nil
}

//CreateThumbnail create video thumbnail
func (f *FFMpeg) CreateThumbnail(filePath string, toFilePath string, width int, height int) error {
	cmd := exec.Command(f.path, "-i", filePath, "-f", "mjpeg", "-vframes", "1", "-y", "-s", fmt.Sprintf("%dx%d", width, height), toFilePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		fmt.Println("cmd.Run() err:", err)
		return ErrCannotCreateThumbnail
	}
	if !fileExists(toFilePath) {
		fmt.Println("toFilePath not exists")
		return ErrCannotCreateThumbnail
	}
	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
