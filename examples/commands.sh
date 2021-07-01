FFMPEG_SRC="./../../libs/ffmpeg/linux/ffmpeg" go test ffmpeg -v -count=1
JSON_CONFIG_PATH="./../../testData.json" go test ffmpeg -v -count=1
FFMPEG_SRC="./../../libs/ffmpeg/linux/ffmpeg" TEST_FILES_SRC="./../../testSrc/" go test ffmpeg -v -count=1