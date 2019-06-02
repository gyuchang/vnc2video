package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
	vnc "vnc2video"
	"vnc2video/encoders"
	log "github.com/sirupsen/logrus"
)

func main() {
	framerate := 10
	speedupFactor := 3.6
	fastFramerate := int(float64(framerate) * speedupFactor)

	if len(os.Args) <= 1 {
		log.Errorf("please provide a fbs file name")
		return
	}
	if _, err := os.Stat(os.Args[1]); os.IsNotExist(err) {
		log.Errorf("File doesn't exist", err)
		return
	}
	encs := []vnc.Encoding{
		&vnc.RawEncoding{},
		&vnc.TightEncoding{},
		&vnc.CopyRectEncoding{},
		&vnc.ZRLEEncoding{},
	}

	fbs, err := vnc.NewFbsConn(
		os.Args[1],
		encs,
	)
	if err != nil {
		log.Error("failed to open fbs reader:", err)
		//return nil, err
	}

	//launch video encoding process:
	ffmpeg := "/usr/local/bin/ffmpeg"
	//vcodec := &encoders.MJPegImageEncoder{Quality: 60 , Framerate: framerate}
	vcodec := &encoders.Encoder{
		BinPath: ffmpeg,
		Cmd: exec.Command(
			ffmpeg,
			"-f", "image2pipe",
			"-vcodec", "ppm",
			"-r", "15",
			"-an",  // no audio
			"-y",
			"-i", "-",
			"-vcodec", "libx264", //"libvpx",//"libvpx-vp9"//"libx264"
			"-threads", "8",
			"-preset", "fast",
			"-g", "250",
			"-crf", "37",
			"./output.mp4",
		),
	}
	//vcodec := &encoders.DV8ImageEncoder{}
	//vcodec := &encoders.DV9ImageEncoder{}
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	log.Debugf("current dir: %s", dir)
	go vcodec.Run()

	//screenImage := image.NewRGBA(image.Rect(0, 0, int(fbs.Width()), int(fbs.Height())))
	screenImage := vnc.NewVncCanvas(int(fbs.Width()), int(fbs.Height()))
	screenImage.DrawCursor = false

	for _, enc := range encs {
		myRenderer, ok := enc.(vnc.Renderer)

		if ok {
			myRenderer.SetTargetImage(screenImage)
		}
	}

	go func() {
		frameMillis := (1000.0 / float64(fastFramerate)) - 2 //a couple of millis, adjusting for time lost in software commands
		frameDuration := time.Duration(frameMillis * float64(time.Millisecond))
		//log.Error("milis= ", frameMillis)

		for {
			timeStart := time.Now()

			vcodec.Encode(screenImage.Image)
			timeTarget := timeStart.Add(frameDuration)
			timeLeft := timeTarget.Sub(time.Now())
			//.Add(1 * time.Millisecond)
			if timeLeft > 0 {
				time.Sleep(timeLeft)
				//log.Error("sleeping= ", timeLeft)
			}
		}
	}()

	msgReader := vnc.NewFBSPlayHelper(fbs)
	//loop over all messages, feed images to video codec:
	for {
		_, err := msgReader.ReadFbsMessage(true, speedupFactor)
		//vcodec.Encode(screenImage.Image)
		if err != nil {
			os.Exit(-1)
		}
		//vcodec.Encode(screenImage)
	}
}
