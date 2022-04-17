// before you run:
// make sure you have ffmpeg
//   either install ffmpeg yourself, or have the ffmpeg[.exe] executable placed in the same directory as this go program
//
// To execute:
// go run blurfaces.go [inputvideo.mp4] [outputvideo.mp4]

package main

import (
    "fmt"
    "bytes"
    "strings"
    "sync"
    "os"
    "os/exec"
    "io/ioutil"
)

func worker(filename string) {
    fmt.Println("Start converting segment", filename)
    cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-vf", "hue=s=0", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "copy", "temp/_"+filename)
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd1.Stdout = &stdout
    cmd1.Stderr = &stderr
    if (cmd1.Run() != nil) {
        fmt.Fprintln(os.Stderr, stderr.String())
    }
    
    fmt.Println("Done converting segment", filename)
    
    err := os.Remove("temp/"+filename)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
    err = os.Rename("temp/_"+filename, "temp/"+filename)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
}

func main() {
    filename := os.Args[1]
    cmd1 := exec.Command("ffmpeg", "-y", "-i", filename, "-f", "segment", "-segment_list", "temp/list.ffcat", "-segment_time", "10", "-reset_timestamps", "1", "-c", "copy", "temp/output_%03d.mp4")
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd1.Stdout = &stdout
    cmd1.Stderr = &stderr
    if (cmd1.Run() != nil) {
        fmt.Fprintln(os.Stderr, stderr.String())
    }
    
    var wg sync.WaitGroup
    items, _ := ioutil.ReadDir("./temp")
    for _, item := range items {
        name := item.Name()
        if !strings.HasPrefix(name,"_") && strings.HasSuffix(name, ".mp4") {
            wg.Add(1)
            go func() {
                defer wg.Done()
                worker(name)
            }()
        }
    }
    wg.Wait()
    
    cmd3 := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", "temp/list.ffcat", "-c", "copy", "output.mp4")
    stdout.Reset()
    stderr.Reset()
    cmd3.Stdout = &stdout
    cmd3.Stderr = &stderr
    if (cmd3.Run() != nil) {
        fmt.Fprintln(os.Stderr, stderr.String())
    }
    
    items, _ = ioutil.ReadDir("./temp")
    for _, item := range items {
        name := item.Name()
        err := os.Remove("temp/"+name)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        }
    }
    
    fmt.Println("Done")
}