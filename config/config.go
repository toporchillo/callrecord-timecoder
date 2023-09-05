package config

var AudioStorage = "sampledata/audio"
var OutputStorage = "sampledata/metadata"

var TempfileDir = "sampledata/tmp" //В этой папке все файды будут удалены
var ChunkSeconds = 3 //Длинна аудиофрагмента в скундах для анализа
var GlueTimecodeLength = 3 //Убирать метки тишины, которые короче GlueTimecodeLength секунд (д/б кратным chunkSeconds)

var SilenceLevel float32 = 0.501 //Уровень сигнала ниже которого мы считаем, что там тишина
var WhitenoizeAutocorrelationLevel float64 = 1e-5 //Автокорреляция сигнала ниже которой мы считаем, что там шум

var FfmpegBinPath = "ffmpeg/bin/ffmpeg"
var FfprobeBinPath = "ffmpeg/bin/ffprobe"
