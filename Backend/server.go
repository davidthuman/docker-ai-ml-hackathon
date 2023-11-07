package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MB

type Word struct {
	Word        string
	Start       float64
	End         float64
	Probability float64
}

func (w *Word) StartTimeString() string {
	return fmt.Sprintf("%02.f:%02.f:%06.2f", math.Floor(w.Start/3600), math.Floor(math.Mod(w.Start, 36000)/60), math.Mod(w.Start, 60))
}

func (w *Word) StartIdString() string {
	return fmt.Sprintf("%.1f", (w.Start*5)/5)
}

func (w *Word) StartIdInt() int {
	return int((w.Start * 5) / 5)
}

func (w *Word) ConfColor() string {
	f := (w.Probability - 0.0) / (1.0 - 0.0)
	red := (f*(0-1) + 1) * 255
	green := (f*(1-0) + 0) * 255
	blue := (f*(0-0) + 0) * 255

	return fmt.Sprintf("#%02x%02x%02x", int(red), int(green), int(blue))
}

type Segment struct {
	Id               int
	Seek             float64
	Start            float64
	End              float64
	Text             string
	Tokens           []int
	Temperature      float64
	AvgLogProb       float64
	CompressionRatio float64
	NoSpeechProb     float64
	Words            []Word
}

type TextResult struct {
	Text     string
	Segments []Segment
	Language string
}

type SpeakerText struct {
	Speaker string
	Result  TextResult
}

type TranscriptPage struct {
	Title      string
	FilePath   string
	Transcript []SpeakerText
}

func homeHandler(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie("userId")
	if err != nil {
		w.Header().Add("HX-Redirect", "/signup/")
		return
	}

	db, err := getDb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user User
	db.First(&user, cookie.Value)

	var audios []Audio
	db.Where("user_id = ?", cookie.Value).Find(&audios)

	userHome := struct {
		User   User
		Audios []Audio
	}{User: user, Audios: audios}

	t, err := template.ParseFiles("./html/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	t.Execute(w, userHome)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {

	// Check for user cookie
	cookie, err := r.Cookie("userId")
	if err != nil {
		w.Header().Add("HX-Redirect", "/signup/")
		return
	}

	// Check request method for POST verb
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check file size
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		http.Error(w, "The uploaded file is too big. Please choose a file that is less than 1MB in size.", http.StatusBadRequest)
		return
	}

	// The argument to FormFile must match the name attribute of the file input on the frontend
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Get the bytes of the file type
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check the file type
	filetype := http.DetectContentType(buff)
	if !strings.Contains(filetype, "audio") {
		http.Error(w, filetype+" format is not allowed. Please upload an audio file.", http.StatusBadRequest)
		return
	}

	// Reset file reader location
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Connect to database
	db, err := getDb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get user ID
	id_i, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	audio := Audio{Name: r.FormValue("name"), Path: "", TranscriptPath: "", UserID: uint(id_i)}
	db.Create(&audio)

	// Make audio directory
	err = os.MkdirAll(fmt.Sprintf("./uploads/%s/%d", cookie.Value, audio.ID), os.ModePerm)
	if err != nil {
		db.Delete(&audio)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Create a new file in the uploads directory
	dst, err := os.Create(fmt.Sprintf("./uploads/%s/%d/input%s", cookie.Value, audio.ID, filepath.Ext(fileHeader.Filename)))
	if err != nil {
		db.Delete(&audio)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	db.Model(&audio).Update("Path", dst.Name())

	// Copy the uploaded file to the filesystem at the specified destination
	_, err = io.Copy(dst, file)
	if err != nil {
		db.Delete(&audio)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<div class=\"flex gap-2\"><div>%s</div><buttom hx-post=\"/transcribe/%d\" hx-swap=\"outerHTML\" class=\"border-2 rounded-md px-2 hover:border-[#1D63ED]\"> Generate Transcript </buttom></div>", audio.Name, audio.ID)
}

func uploadsHandler(w http.ResponseWriter, r *http.Request) {

	body, err := os.ReadFile("./" + r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(body)
}

func transcribeHandler(w http.ResponseWriter, r *http.Request) {

	// Check for user cookie
	cookie, err := r.Cookie("userId")
	if err != nil {
		w.Header().Add("HX-Redirect", "/signup/")
		return
	}

	splits := strings.Split(r.URL.Path, "/")
	audioId := splits[len(splits)-1]

	dockerTranscribe(fmt.Sprintf("/uploads/%s/%s", cookie.Value, audioId))

	// Connect to database
	db, err := getDb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var audio Audio
	db.First(&audio, audioId)
	db.Model(&audio).Update("TranscriptPath", fmt.Sprintf("./uploads/%s/%s/transcript.json", cookie.Value, audioId))

	fmt.Fprintf(w, "<a href=\"/transcript/%s\" class=\"border-2 rounded-md px-2 hover:border-[#1D63ED]\"> View Transcript </a>", audioId)
}

func trascriptHanlder(w http.ResponseWriter, r *http.Request) {

	// Check for user cookie
	_, err := r.Cookie("userId")
	if err != nil {
		w.Header().Add("HX-Redirect", "/signup/")
		return
	}

	splits := strings.Split(r.URL.Path, "/")
	audioId := splits[len(splits)-1]

	// Connect to database
	db, err := getDb()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var audio Audio
	db.First(&audio, audioId)

	data, err := os.ReadFile("./" + audio.TranscriptPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var speakerText []SpeakerText
	err = json.Unmarshal(data, &speakerText)
	_, i := utf8.DecodeLastRuneInString(audio.Path)

	transcriptPage := TranscriptPage{
		Title:      audio.Name,
		FilePath:   audio.Path[i:],
		Transcript: speakerText,
	}

	t, err := template.ParseFiles("./html/view.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, transcriptPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func signupHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, err := template.ParseFiles("./html/signup.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	if r.Method == "POST" {

		db, err := getDb()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := User{Email: r.FormValue("email"), Name: r.FormValue("name")}
		db.Create(&user)

		// Create the user_id folder
		err = os.MkdirAll(fmt.Sprintf("./uploads/%s", fmt.Sprint(user.ID)), os.ModePerm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Add("HX-Redirect", "/home/")
		cookie := http.Cookie{
			Name:     "userId",
			Value:    fmt.Sprint(user.ID),
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, &cookie)
	}
}

func main() {
	http.HandleFunc("/signup/", signupHandler)         // For creating a new user
	http.HandleFunc("/home/", homeHandler)             // For user navigation
	http.HandleFunc("/upload/", uploadHandler)         // For uploading an audio file
	http.HandleFunc("/uploads/", uploadsHandler)       //
	http.HandleFunc("/transcript/", trascriptHanlder)  //
	http.HandleFunc("/transcribe/", transcribeHandler) //

	// Create the uploads folder if it doesn't already exist
	err := os.MkdirAll("./database", os.ModePerm)
	if err != nil {
		panic(err)
	}
	// Create the uploads folder if it doesn't already exist
	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		panic(err)
	}
	initDb()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
