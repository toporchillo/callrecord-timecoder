package main

import (
    "log"
    "os"
    "path/filepath"
    "callrecord-timecoder/config"
    "callrecord-timecoder/mediaconvert"
    "callrecord-timecoder/wavanalyser"
    "github.com/mjibson/go-dsp/wav"
    "fmt"
    //"errors"
    "encoding/json"
    "math"
    "regexp"
    "time"
    "strings"
)

const SILENCE byte = 1
const VOICE byte = 10

func logIt(v ...any) {
        fmt.Println(v)
		log.Println(v)
}

func main() {
    logfile, err := os.OpenFile("log/app.log", os.O_APPEND|os.O_WRONLY, 0666)
    if err != nil {
        fmt.Println("Open/create log file error")
        panic("Open/create log file error")
    }
    log.SetOutput(logfile)
    defer logfile.Close()

    defer cleanUp()

	logIt("Process dir:", config.AudioStorage, "ChunkSeconds:", config.ChunkSeconds, "GlueTimecodeLength:", config.GlueTimecodeLength, "SilenceLevel:", config.SilenceLevel, "WhitenoizeAutocorrelationLevel:", config.WhitenoizeAutocorrelationLevel);
    processStorageDir(config.AudioStorage)
}

func processStorageDir(dir string) (error) {
	files, err := os.ReadDir(dir)

	if err != nil {
		logIt(dir, " read dir error: ", err)
		return err
	}

	for _, file := range files {
	    if file.IsDir() {
	        processStorageDir2(dir + "/" + file.Name())
	    } else {
		    ret, err := analyzeAudioFile(dir + "/" + file.Name())
		    if err == nil {
                logIt(dir + "/" + file.Name(), "OK", string(ret))

                dest_dir := filepath.Dir(config.OutputStorage + "/" + strings.Replace(dir, config.AudioStorage, "", 1) + "/" + file.Name())
                err = os.MkdirAll(dest_dir, 0770)
                if err != nil {
            		logIt(dest_dir, " create output dir error: ", err)
                } else {
                    f,_ := os.Create(dest_dir + "/" + file.Name() + ".json")
                    f.WriteString(string(ret))
                    defer f.Close()
                }
            }

		}
	}
    return nil
}

func processStorageDir2(dir string) (error) {
    return processStorageDir(dir)
}

/*
Очистка временные wav-файлов.
Из-за того, что ffmpeg долго освобождает файл, приходится делать Sleep
*/
func cleanUp() {
    time.Sleep(5*time.Second)
    var validFn = regexp.MustCompile(`^tc\-[0-9]+\.wav$`)
	files, _ := os.ReadDir(config.TempfileDir)
	for _, file := range files {
        if (validFn.MatchString(file.Name())) {
            err := os.Remove(config.TempfileDir + "/" + file.Name())
            if err != nil {
                fmt.Println(file.Name(), "  clean up temp file error: ", err)
            }
        }
	}
}

func analyzeAudioFile(filepath string) ([]byte, error) {
    file, err := os.CreateTemp(config.TempfileDir, "tc-*.wav")

    if err != nil {
        logIt(filepath, " create temp file error: ", err)
        return nil, err
    }

    defer os.Remove(file.Name())

    err =  mediaconvert.Audio2Wav(filepath, file.Name())

    if err != nil {
        logIt(filepath, " audio convert error: ", err)
        return nil, err
    }

	wavdata, err := wav.New(file)

	if err != nil {
		logIt(filepath, " read wav file error: ", err)
        return nil, err
	}

    /*
    fmt.Printf("Duration: %v\n", wavdata.Duration)
    fmt.Printf("Samples: %v\n", wavdata.Samples)
    fmt.Printf("BitsPerSample: %v\n", wavdata.Header.BitsPerSample)
    fmt.Printf("NumChannels: %v\n", wavdata.Header.NumChannels)
    fmt.Printf("BlockAlign: %v\n", wavdata.Header.BlockAlign)
    fmt.Printf("SampleRate: %v\n", wavdata.Header.SampleRate)
    fmt.Printf("ByteRate: %v\n", wavdata.Header.ByteRate)
    */

    //Кол-во float-чисел в одной секунде аудио
    floatrate:= int(wavdata.Header.ByteRate) * 8 / int(wavdata.Header.BitsPerSample)

    i := 0
    j := 0

    var prev byte = 0
    var current byte = 0

    var timecodes []int
    var signals []byte

	for i<wavdata.Samples {
    	floats, _ := wavdata.ReadFloats(int(math.Round(float64(floatrate) * float64(config.ChunkSeconds))))
	    i+= floatrate

        if wavanalyser.IsSilence(floats) {
            current = SILENCE
        } else {
        	chunk := mediaconvert.MixChannels(floats, int(wavdata.Header.NumChannels))

            if (wavanalyser.IsWhiteNoize(chunk)) {
                current = SILENCE
            } else {
                current = VOICE
            }
        }

        if prev != current {
            timecodes = append(timecodes, int(math.Round(float64(j) * float64(config.ChunkSeconds))))
            signals = append(signals, current)
        }

        prev = current
	    j++
	}

    ret_tc, ret_s := glueTimecodes(timecodes, signals, config.GlueTimecodeLength)

    ret, err := timecodesToJson(ret_tc, ret_s)

	if err != nil {
		logIt(filepath, " JSON.stringify error ", err)
        return nil, err
	}

	return ret, nil
}

/**
Склеивает фрагменты с таймкодами, исключая фрагменты короче len
**/
func glueTimecodes(timecodes []int, signals []byte, len int) ([]int, []byte) {
    prev_k := timecodes[0]
    prev_v := signals[0]
    var temp_tc []int
    var temp_s []byte
    var ret_tc []int
    var ret_s []byte

    for i, k := range timecodes {
        v := signals[i]
        if (k==0 || (prev_v == SILENCE && (k - prev_k) >= len) || prev_v == VOICE) {
            temp_tc = append(temp_tc, prev_k)
            temp_s = append(temp_s, prev_v)
        }
        prev_k = k
        prev_v = v
    }

    prev_k = -1
    prev_v = 0
    for i, k := range temp_tc {
        v := temp_s[i]
        if (prev_v != v) {
            ret_tc = append(ret_tc, k)
            ret_s = append(ret_s, v)
        }
        prev_k = k
        prev_v = v
    }
    return ret_tc, ret_s
}

func timecodesToJson(ret_tc []int, ret_s []byte) ([]byte, error) {
    type Timecode struct {
        O int
        S  string
    }

    var ret []Timecode
    for i, k := range ret_tc {
        v := ret_s[i]
        sound := "V"
        if (v == SILENCE) {
            sound = "S"
        }
        item := Timecode{O: k, S: sound}
        ret = append(ret, item)
    }

    return json.Marshal(ret)
}