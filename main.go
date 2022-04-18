package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
  "bytes"
  "strings"
  "sync"
  "os"
  "os/exec"
  "io/ioutil"
  "strconv"
  "image"
  "gocv.io/x/gocv"
)

type filterName struct {
  name string
}

var f = filterName{name: "Grayscale"}

// Filter functions

func worker(fileIn string, filtername string) {
    fmt.Println("Start converting segment", fileIn)
		cmd1 := exec.Command("ffmpeg", "-y", "-i", "temp/"+fileIn, "-vf", "hue=s=0", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "copy", "temp/_"+fileIn)
		switch filtername {
		case "Rotate":
			cmd1 = exec.Command("ffmpeg", "-y", "-i", "temp/"+fileIn, "-c", "copy", "-metadata:s:v:0", "rotate=180", "-pix_fmt", "yuv420p", "temp/_"+fileIn)
		case "Brighten":
			cmd1 = exec.Command("ffmpeg", "-y", "-i", "temp/"+fileIn, "-vf", "eq=brightness=0.06:saturation=2", "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "copy", "temp/_"+fileIn)
		case "DoubleSpeed":
			cmd1 = exec.Command("ffmpeg", "-y", "-i", "temp/"+fileIn, "-filter:v", "setpts=0.5*PTS", "temp/_"+fileIn)
		case "HalfSpeed":
			cmd1 = exec.Command("ffmpeg", "-y", "-i", "temp/"+fileIn, "-filter:v", "setpts=2.0*PTS", "temp/_"+fileIn)
		}
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd1.Stdout = &stdout
    cmd1.Stderr = &stderr
    if (cmd1.Run() != nil) {
        fmt.Fprintln(os.Stderr, stderr.String())
    }

    fmt.Println("Done converting segment", fileIn)

    err := os.Remove("temp/"+fileIn)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
    err = os.Rename("temp/_"+fileIn, "temp/"+fileIn)
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }
}


func run() {
  xmlFile := "haarcascade_frontalface_default.xml"
  if f.name == "Blurfaces" {
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

	} else {
		cmd1 := exec.Command("ffmpeg", "-y", "-i", fileIn, "-f", "segment", "-segment_list", "temp/list.ffcat", "-segment_time", "10", "-reset_timestamps", "1", "-c", "copy", "temp/output_%03d.mp4")
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
                worker(name, f.name)
            }()
        }
    }
    wg.Wait()

    cmd2 := exec.Command("ffmpeg", "-y", "-f", "concat", "-safe", "0", "-i", "temp/list.ffcat", "-c", "copy", "output.mp4")

		stdout.Reset()
	  stderr.Reset()
	  cmd2.Stdout = &stdout
	  cmd2.Stderr = &stderr
	  if (cmd2.Run() != nil) {
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
}

func filter(value string) {
	f.name = value
}

var fileIn string  // Input File's Address
var fileOut string // output File's Address

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Go Group")
	//var fileIn string
	input := widget.NewButton("Input .mp4 files", func() {
		file_Dialog := dialog.NewFileOpen(
			func(r fyne.URIReadCloser, _ error) {
				fileIn = r.URI().Path()
				log.Println("Input Address is", fileIn)
			}, myWindow)
		file_Dialog.Show()
		// Show file selection dialog.
	})

	output := widget.NewButton("Output Address", func() {
		file_Dialog := dialog.NewFolderOpen(
			func(r fyne.ListableURI, _ error) {
				fileOut = r.Path()
				fmt.Println(fileOut)
			}, myWindow)
		file_Dialog.Show()
		// Show file selection dialog.
		log.Println("Output Address is", fileOut)
	})

	combo := widget.NewSelect([]string{"Grayscale", "Blurfaces", "Rotate",
    "Brighten", "DoubleSpeed", "HalfSpeed"}, filter)
	run := widget.NewButton("Run", run)
	myWindow.SetContent(container.NewVBox(input, combo, output, run))
	myWindow.Resize(fyne.NewSize(400, 400))
	myWindow.ShowAndRun()

}
