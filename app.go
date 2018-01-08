package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
	// "io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
)

type CommandInput struct {
	Command string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello world!")
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	var commandInput CommandInput
	err := json.NewDecoder(r.Body).Decode(&commandInput)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on parsing form: %v", err), http.StatusBadRequest)
		return
	}
	command := commandInput.Command

	err = executeCommand(w, command)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on execute command \"%s\": %v", command, err), http.StatusBadRequest)
		return
	}

	log.Println("Finish:", command)
}

func executeCommand(w http.ResponseWriter, command string) error {
	stderrBuf := new([]byte)
	running := exec.Command("bash", "-c", command)
	stdout, err := running.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := running.StderrPipe()
	if err != nil {
		return err
	}

	err = running.Start()
	if err != nil {
		return err
	}
	log.Println("Running:", command)

	// go func() {
	// 	*stderrBuf, err = ioutil.ReadAll(stderr)
	// 	if err != nil {
	// 		log.Printf("Error on reading stderr: %v", err)
	// 	}
	// }()

	go func() {
		_, err = io.Copy(w, stdout)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error on piping \"%s\": %v", command, err), http.StatusBadRequest)
		}
	}()

	ticker := time.NewTicker(time.Millisecond * 500)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			w.(http.Flusher).Flush()
		}
	}()

	_, err = io.Copy(w, stderr)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error on piping stderr \"%s\": %v", command, err), http.StatusBadRequest)
	}

	err = running.Wait()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error waiting \"%s\": %v", command, err), http.StatusBadRequest)
	}

	log.Println("err:", string(*stderrBuf))
	return nil
}

func main() {
	port := getEnv("PORT", "8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/execute", executeHandler)

	log.Print("listening :", port)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", port), mux))
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
