package bot

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

func convertOgaToMp3(inputFile, outputFile string) error {
    // Check if input file exists
    if _, err := os.Stat(inputFile); os.IsNotExist(err) {
        return fmt.Errorf("input file does not exist: %s", inputFile)
    }
    
    // If no output file is specified, create one with the same name but .mp3 extension
    if outputFile == "" {
        ext := filepath.Ext(inputFile)
        outputFile = inputFile[0:len(inputFile)-len(ext)] + ".mp3"
    }
    
    // Prepare FFmpeg command
    cmd := exec.Command("ffmpeg", "-i", inputFile, "-acodec", "libmp3lame", "-q:a", "2", outputFile)
    
    // Capture standard output and error
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("conversion failed: %v\nOutput: %s", err, output)
    }
    
    slog.Debug("Successfully converted %s to %s\n", inputFile, outputFile)
    return nil
}