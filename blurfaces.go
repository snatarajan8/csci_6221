// before you run:
// 1. make sure you have opencv4 and gocv installed
//    installation instructions: https://gocv.io/getting-started/
// 2. also make sure you have ffmpeg
//    either install ffmpeg yourself, or have the ffmpeg[.exe] executable placed in the same directory as this go program
// 3. you also need "haarcascade_frontalface_default.xml" in the same directory
//    download it yourself and place it in the same directory as this go program
//
// To execute:
// go run blurfaces.go [inputvideo.mp4] [outputvideo.mp4]

package main

import (
    "fmt"
    "bytes"
    "strings"
    "strconv"
    "image"
    "os"
    "os/exec"
    "io/ioutil"
    "gocv.io/x/gocv"
)

func main() {
    if len(os.Args) < 2 {
      fmt.Println("To run:\n\tgo run blurfaces.go [Input File] [Output File]")
      return
    }
    fileIn := os.Args[1]
    fileOut := "output.mp4"
    if len(os.Args) >= 3 {
      fileOut = os.Args[2]
    }
    
    xmlFile := "haarcascade_frontalface_default.xml"
    
    cmd1 := exec.Command("ffprobe", "-v", "error", "-select_streams", "v", "-of", "default=noprint_wrappers=1:nokey=1", "-show_entries", "stream=width,height,r_frame_rate", fileIn)
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd1.Stdout = &stdout
    cmd1.Stderr = &stderr
    if (cmd1.Run() != nil) {
        fmt.Fprintln(os.Stderr, stderr.String())
        return
    }
    
    streamInfo := strings.Fields(stdout.String())
    width,_ := strconv.Atoi(streamInfo[0])
    height,_ := strconv.Atoi(streamInfo[1])
    fpsString := strings.Split(streamInfo[2],"/")
    dividend,_ := strconv.ParseFloat(fpsString[0],64)
    divisor,_ := strconv.ParseFloat(fpsString[1],64)
    framerate := dividend / divisor
    
    videoIn,err := gocv.VideoCaptureFile(fileIn)
    if err != nil {
      fmt.Println("Error opening file: ", fileIn)
      fmt.Println(err)
      return
    }
    defer videoIn.Close()
    
    picture := gocv.NewMat()
    defer picture.Close()
    
    classifier := gocv.NewCascadeClassifier()
    defer classifier.Close()
    
    if !classifier.Load(xmlFile) {
      fmt.Println("Error reading cascade file: ", xmlFile)
      return
    }
    
    writer, err := gocv.VideoWriterFile("temp/_"+fileOut, "avc1", framerate, width, height, true)
    if err != nil {
      fmt.Println("error opening video writer device: ", "temp/_"+fileOut)
      fmt.Println(err)
      return
    }
    
    framecounter := 0
    for {
      framecounter += 1
      
      if ok := videoIn.Read(&picture); !ok {
        fmt.Println("Video File Ended: ", fileIn)
        break
      }
      if picture.Empty() {
        continue
      }
      
      rects := classifier.DetectMultiScale(picture)
      for _,r := range rects {
        imgFace := picture.Region(r)
        gocv.GaussianBlur(imgFace, &imgFace, image.Pt(75,75), 0, 0, gocv.BorderDefault)
        imgFace.Close()
      }
      
      writer.Write(picture)
      if framecounter % 20 == 0 {
        fmt.Println("Done writing frame number", framecounter)
      }
    }
    writer.Close()
    
    cmd2 := exec.Command("ffmpeg", "-y", "-i", "temp/_"+fileOut, "-i", fileIn, "-map", "0:v:0", "-map", "1:a:0", "-c", "copy", fileOut)
    stdout.Reset()
    stderr.Reset()
    cmd2.Stdout = &stdout
    cmd2.Stderr = &stderr
    if (cmd2.Run() != nil) {
        fmt.Fprintln(os.Stderr, stderr.String())
    }
    
    items, _ := ioutil.ReadDir("./temp")
    for _, item := range items {
        name := item.Name()
        err := os.Remove("temp/"+name)
        if err != nil {
            fmt.Fprintln(os.Stderr, err)
        }
    }
    
    fmt.Println("Done")
    return
}
