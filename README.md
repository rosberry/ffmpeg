# FFMPEG

Library wrapper over ffmpeg.



Init:
```golang
ff := ffmpeg.New().SetPath("lib/ffmpeg")
```


*Functions:*

Get duration with encoding
```golang
func (f *FFMpeg) Duration(filePath string) (*time.Duration, error)
```

Get duration without encoding (not for all codecs)
```golang
func (f *FFMpeg) SimpleDuration(filePath string) (*time.Duration, error)
```

Trim video
```golang
func (f *FFMpeg) TrimVideo(filePath string, toFilePath string, start int, duration int) error
```

Get file bitrate
```golang
func (f *FFMpeg) Bitrate(filePath string) (*string, error)
```

Get file title
```golang
func (f *FFMpeg) Title(filePath string) (*string, error)
```

Get file size
```golang
func (f *FFMpeg) Size(filePath string) (uint, uint, error)
```

Create video thumbnail
```golang
func (f *FFMpeg) CreateThumbnail(filePath string, toFilePath string, width int, height int) error
```