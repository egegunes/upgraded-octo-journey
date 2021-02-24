package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

type task struct {
	Name     string
	Duration int
}

const MAX = 30

func main() {
	wait := make(chan task, 100)
	run := make(chan task, 20)
	done := make(chan task, 20)

	var waiting = struct {
		sync.RWMutex
		m map[string]task
	}{m: make(map[string]task)}

	var running = struct {
		sync.RWMutex
		m map[string]task
	}{m: make(map[string]task)}

	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, fmt.Sprintf("%s not allowed", r.Method), 405)
			return
		}

		waiting.RLock()
		fmt.Fprintf(w, "waiting %d\n", len(waiting.m))
		for k, v := range waiting.m {
			fmt.Fprintln(w, k, v.Duration)
		}
		waiting.RUnlock()

		running.RLock()
		fmt.Fprintf(w, "\nrunning %d\n", len(running.m))
		for k, v := range running.m {
			fmt.Fprintln(w, k, v.Duration)
		}
		running.RUnlock()

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, fmt.Sprintf("%s not allowed", r.Method), 405)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("can't read body: %v", err), 400)
			return
		}
		defer r.Body.Close()

		d := make(map[string]int)
		if err := json.Unmarshal(body, &d); err != nil {
			http.Error(w, fmt.Sprintf("can't umarshal body to json: %v", err), 400)
			return
		}

		for k, v := range d {
			wait <- task{Name: k, Duration: v}
		}

		w.WriteHeader(http.StatusCreated)
	})

	go func() {
		for {
			t := <-wait
			fmt.Printf("waiting: %s\n", t.Name)

			waiting.Lock()
			waiting.m[t.Name] = t
			waiting.Unlock()

			run <- t

		}
	}()

	go func() {
		for {
			if len(running.m) >= MAX {
				continue
			}

			t := <-run

			running.RLock()
			if _, ok := running.m[t.Name]; ok {
				fmt.Printf("warning: %s is already running\n", t.Name)
				running.RUnlock()
				continue
			}
			running.RUnlock()

			waiting.Lock()
			delete(waiting.m, t.Name)
			waiting.Unlock()

			running.Lock()
			running.m[t.Name] = t
			running.Unlock()

			go func() {
				fmt.Printf("running: %s\n", t.Name)
				time.Sleep(time.Duration(t.Duration) * time.Millisecond)
				done <- t
			}()

		}
	}()

	go func() {
		for {
			t := <-done

			running.Lock()
			delete(running.m, t.Name)
			running.Unlock()

			fmt.Printf("done: %s\n", t.Name)
		}
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
