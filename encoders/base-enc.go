package encoders

import (
	log "github.com/sirupsen/logrus"
	"image"
	"io"
	"os"
	"os/exec"
)

type Encoder struct {
	BinPath   string
	Cmd       *exec.Cmd
	input     io.WriteCloser
	closed    bool
	Framerate int
}

func (enc *Encoder) Run() error {
	enc.Cmd.Stdout = os.Stdout
	enc.Cmd.Stderr = os.Stderr

	encInput, err := enc.Cmd.StdinPipe()
	enc.input = encInput
	if err != nil {
		log.Fatalf("can't get input pipe from", enc.BinPath)
		panic(err)
	}

	log.Infof("launching binary: %s", enc.BinPath)
	err = enc.Cmd.Run()
	log.Infof("exit.", enc.BinPath)
	if err != nil {
		log.Errorf("error while launching %s: %v", enc.BinPath, enc.Cmd.Args)
		return err
	}
	return nil
}
func (enc *Encoder) Encode(img image.Image) {
	if enc.input == nil || enc.closed {
		return
	}

	err := encodePPM(enc.input, img)
	if err != nil {
		log.Errorf("error while encoding image: %v", err)
	}
}

func (enc *Encoder) Close() {
	enc.closed = true
	//enc.input.Close()
	//enc.Cmd.Process.Signal(os.Interrupt)
	//enc.Cmd.Process.Release()
}
