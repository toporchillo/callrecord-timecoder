package mediaconvert

import (
	"log"
	ffmpeg "github.com/floostack/transcoder/ffmpeg"
	"callrecord-timecoder/config"
)

//Convert input audio file to output PCM WAV file
func Audio2Wav(input string, output string) error {

	format := "wav"
	overwrite := true

	opts := ffmpeg.Options{
		OutputFormat: &format,
		Overwrite:    &overwrite,
	}

	ffmpegConf := &ffmpeg.Config{
		FfmpegBinPath:   config.FfmpegBinPath,
		FfprobeBinPath:  config.FfprobeBinPath,
		ProgressEnabled: true,
	}

	progress, err := ffmpeg.
		New(ffmpegConf).
		Input(input).
		Output(output).
		//WithOptions(opts).
		Start(opts)

	if err != nil {
		return err
	}

	for msg := range progress {
		log.Printf("%+v", msg)
	}
	return nil
}

func MixChannels(arr []float32, channels int) ([]float64) {

    ret := make([]float64, len(arr)/channels)
    j:= 0
    i:= 0

    for (j*channels + i) < len(arr) {

        var avg float32
        avg = 0

        for i<channels {
            avg+= arr[j*channels + i]
            i++
        }
        ret[j] = float64(avg/float32(channels))
        i = 0
        j++
    }
    return ret
}