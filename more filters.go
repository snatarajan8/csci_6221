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
    //cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-vf", "hue=s=0", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "copy", "temp/_"+filename)
    //cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-c", "copy", "-metadata:s:v:0", "rotate=180", "-pix_fmt", "yuv420p", "temp/_"+filename) 'rotate 180°'
    //cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-vf", "rotate=angle=-20*PI/180:fillcolor=brown", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "copy", "temp/_"+filename) 'a more vivid rotate method'
    //cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-vf", "eq=brightness=0.06:saturation=2", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "copy", "temp/_"+filename) 'brighten the vid'
    //cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-i", "logo.png", "-filter_complex", "[1:v]scale=176:144[logo];[0:v][logo]overlay=x=0:y=0", "temp/_"+filename) 'add logo, a file call logo.png is needed'
    //cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-filter:v", "setpts=0.5*PTS", "temp/_"+filename) 'Double speed playback'
    //cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+filename, "-filter:v", "setpts=2.0*PTS", "temp/_"+filename) 'slow play'
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