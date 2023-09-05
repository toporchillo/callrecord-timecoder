package wavanalyser
import (
	"math"
    "github.com/mjibson/go-dsp/fft"
    "math/cmplx"
    "github.com/montanaflynn/stats"
    //"github.com/mjibson/go-dsp/spectral"
    "callrecord-timecoder/config"
)

func IsSilence(floats []float32) bool {
    for i:= range floats {
        if (floats[i] > config.SilenceLevel) {
            return false
        }
    }
    return true
}

func IsWhiteNoize(floats []float64) bool {
    var fftres []float64

    for _, complex := range fft.FFTReal(floats) {
        fftres = append(fftres, cmplx.Abs(complex))
    }
    a, _ := stats.AutoCorrelation(fftres, 2)
    return a < config.WhitenoizeAutocorrelationLevel
}

/*
var Sopts = spectral.PwelchOptions{NFFT: 256}

func IsNoize(floats []float64, sampleRate float64) bool {
    spectr, freqs := spectral.Pwelch(floats, sampleRate, &Sopts)

    var spectr_filtered []float64

    for i:= range freqs {
        if freqs[i] > 100 && freqs[i] < 1000 {
            spectr_filtered = append(spectr_filtered, spectr[i]*10000000)
        }
    }
    a, _ := stats.AutoCorrelation(spectr_filtered, 2)
    fmt.Printf("Pwelch AutoCorrelation: %v\n", a)
    return a < WhitenoizeAutocorrelationLevel
}

func GetDiff(Pxx []float64, freqs []float64) float64 {
    var arr []float64

    for i:= range freqs {
        if freqs[i] > 100 && freqs[i] < 1000 {
            arr = append(arr, Pxx[i]*10000000)
        }
    }
    diff := stdev(arr)
    return diff
}
*/

/**
Среднее арифметическое
**/
func avg(floats []float64) float64 {
    var avg float64 = 0

    for i:= range floats {
        avg+= floats[i]
    }
    return avg/float64(len(floats))
}

/**
Среднеквадратичное отклонение
**/
func stdev(floats []float64) float64 {
    var diff float64 = 0

    avg := avg(floats)

    for i:= range floats {
        diff+= (floats[i] - avg)*(floats[i] - avg)
    }
    return math.Sqrt(diff/float64(len(floats)-1))
}
