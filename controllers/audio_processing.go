package controllers

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/gin-gonic/gin"
)

// This command is used to update and install the FFmpeg package on a system that uses
// RUN apt-get -y update && apt-get -y upgrade && apt-get install -y --no-install-recommends ffmpeg

// slowed: .85
// fast: 1.15

// var filter_path string = "./LexiconPCM90_Halls/"
// transform("carelesswhisper", "https://www.youtube.com/watch?v=oKtdps9Lm7A", filter_path+"CUSTOM_pump_verb.WAV")

type audio_request struct {
	Url string `json:"url"`
}

func Init_audio_processing(c *gin.Context) {
	// youtube link
	var body audio_request
	if err := c.BindJSON(&body); err != nil {
		c.String(400, "Invalid request body")
		return
	}
	c.JSON(200, gin.H{
		"message": body.Url,
	})
}

func transform(fileName string, url string, filter string) {
	// download video
	if !getMP3FromYotube(url, fileName) {
		return // error downloading mp3
	}

	// add reverb
	fileName = fileName + ".mp3"
	fileNameRev := "reverb_" + fileName
	fmt.Println("Adding reverb...")
	reverbCommand := exec.Command("ffmpeg", "-i", fileName, "-i", filter, "-filter_complex",
		"[0] [1] afir=dry=10:wet=10 [reverb]; [0] [reverb] amix=inputs=2:weights=10 1", fileNameRev)
	reverbOutput, err := reverbCommand.CombinedOutput()
	if logErr(err, reverbOutput) || !deleteFile(fileName) {
		fmt.Println("ERR: Found in reverbing process or deleting excess file.")
		return
	}

	// alter pitch
	fileNamePit := "pitch_" + fileNameRev
	fmt.Println("Lowering pitch...")
	pitchCommand := exec.Command("ffmpeg", "-i", fileNameRev, "-af", "asetrate=44100*0.85,aresample=44100", fileNamePit)
	pitchOutput, err := pitchCommand.CombinedOutput()
	if logErr(err, pitchOutput) || !deleteFile(fileNameRev) {
		fmt.Println("ERR: Found in altering pitch process or deleting excess file.")
		return
	}

	fmt.Println(getThumbnail(url))
	fmt.Println(getVideoLength(fileNamePit))
	fmt.Println("Complete!")
}

func getMP3FromYotube(url string, fileName string) bool {
	fileNameMP4 := fileName + ".mp4"
	fileNameMP3 := fileName + ".mp3"

	// Uses youtube-dl exec on machine to download videos from youtube
	fmt.Println("Downloaded mp4 file...")
	downloadCommand := exec.Command("youtube-dl", "-f", "best", "-o", fileNameMP4, url)
	downloadOutput, err := downloadCommand.CombinedOutput()
	if logErr(err, downloadOutput) {
		return false
	}

	// converts mp4 to mp3 using ffmpeg
	fmt.Println("Converting mp4 to mp3 file...")
	convertCommand := exec.Command("ffmpeg", "-i", fileNameMP4, fileNameMP3)
	convertOutput, err := convertCommand.CombinedOutput()
	if logErr(err, convertOutput) {
		return false
	}

	// removes uneeded mp4 file
	fmt.Println("Removing mp4 file...")
	if !deleteFile(fileNameMP4) {
		return false
	}

	fmt.Println("Successfully downloaded and converted YouTube video to MP3!")
	return true
}

func getThumbnail(url string) string {
	// compile regex expression
	regex := regexp.MustCompile(`v=([^&]+)`)

	// find instance within vid url
	videoID := regex.FindStringSubmatch(url)[1]

	return "https://img.youtube.com/vi/" + videoID + "/maxresdefault.jpg"
}

func getVideoLength(fileName string) string {
	readCommand := exec.Command("ffmpeg", "-i", fileName)
	readOutput, _ := readCommand.CombinedOutput()

	// search for video duration
	regex := regexp.MustCompile(`Duration:\s([0-9]{2}:[0-9]{2}:[0-9]{2}\.[0-9]{2})`)

	// find instance within vid url
	videoDuration := regex.FindStringSubmatch(string(readOutput))[1]

	return videoDuration
}

func deleteFile(fileName string) bool {
	deleteCommand := exec.Command("rm", "-r", fileName)
	deleteOutput, err := deleteCommand.CombinedOutput()
	return !logErr(err, deleteOutput)
}

func logErr(err error, output []byte) bool {
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	return err != nil
}