package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/shirou/gopsutil/process"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type TimerD struct {
	timer        *time.Timer
	tickDuration time.Duration
	remaining    time.Duration
	lastStarted  time.Time
	running      bool
}

type ProgramInfo struct {
	PID  int32
	Name string
}

type MoniData struct {
	Moni string `json:"moni"`
}

func (a *App) StartDualTimer(topBarTick, bottomBarTick int) {
	if a.topTimer.timer != nil || a.bottomTimer.timer != nil {
		a.StopDualTimer()
	}

	a.topTimer.tickDuration = time.Duration(topBarTick) * time.Millisecond
	a.topTimer.remaining = a.topTimer.tickDuration

	a.bottomTimer.tickDuration = time.Duration(bottomBarTick) * time.Millisecond
	a.bottomTimer.remaining = a.bottomTimer.tickDuration

	// Start top timer goroutine
	if !a.topTimer.running {
		a.topTimer.running = true
		go a.startTopTimer()
	}

	// Start bottom timer goroutine
	if !a.bottomTimer.running {
		a.bottomTimer.running = true
		go a.startBottomTimer()
	}

	a.pausedTimers = false
}

// Separate function to start the top timer and handle its events
func (a *App) startTopTimer() {
	a.topTimer.timer = time.AfterFunc(a.topTimer.remaining, func() {
		runtime.EventsEmit(a.ctx, "topBarTick", time.Now().Unix())
		a.topTimer.remaining = a.topTimer.tickDuration
		if a.topTimer.running {
			a.startTopTimer() // Reschedule the timer
		}
	})
}

// Separate function to start the bottom timer and handle its events
func (a *App) startBottomTimer() {
	a.bottomTimer.timer = time.AfterFunc(a.bottomTimer.remaining, func() {
		runtime.EventsEmit(a.ctx, "bottomBarTick", time.Now().Unix())
		a.bottomTimer.remaining = a.bottomTimer.tickDuration
		if a.bottomTimer.running {
			a.startBottomTimer() // Reschedule the timer
		}
	})
}

func (a *App) PauseDualTimer() {
	if a.topTimer.timer != nil && a.topTimer.running {
		a.topTimer.remaining -= time.Since(a.topTimer.lastStarted)
		a.topTimer.timer.Stop()
		a.topTimer.running = false
	}
	if a.bottomTimer.timer != nil && a.bottomTimer.running {
		a.bottomTimer.remaining -= time.Since(a.bottomTimer.lastStarted)
		a.bottomTimer.timer.Stop()
		a.bottomTimer.running = false
	}
}

func (a *App) ResumeDualTimer() {
	if !a.topTimer.running {
		a.topTimer.lastStarted = time.Now()
		a.topTimer.running = true
		a.startTopTimer()
	}
	if !a.bottomTimer.running {
		a.bottomTimer.lastStarted = time.Now()
		a.bottomTimer.running = true
		a.startBottomTimer()
	}
}

func (a *App) StopDualTimer() {
	if a.topTimer.timer != nil {
		a.topTimer.timer.Stop()
		a.topTimer.timer = nil
		a.topTimer.running = false
		a.topTimer.remaining = 0
	}
	if a.bottomTimer.timer != nil {
		a.bottomTimer.timer.Stop()
		a.bottomTimer.timer = nil
		a.bottomTimer.running = false
		a.bottomTimer.remaining = 0
	}
}

func (a *App) GetRunningPrograms() ([]ProgramInfo, error) {
	processes, err := process.Processes()
	if err != nil {
		fmt.Println("Error fetching processes:", err)
		return nil, err
	}

	seen := make(map[string]bool)
	var programs []ProgramInfo

	for _, proc := range processes {
		name, err := proc.Name()
		if err == nil && name != "" && name != "[System Process]" {
			if !seen[name] {
				programs = append(programs, ProgramInfo{PID: proc.Pid, Name: name})
				seen[name] = true
			}
		}
	}

	return programs, nil
}

func (a *App) CheckFocused(programs []string) bool {
	focused := robotgo.GetPid()
	for _, name := range programs {
		fpid, _ := robotgo.FindIds(name)
		for _, pid := range fpid {
			if focused == pid {
				return true
			}
		}
	}
	return false
}

func (a *App) SaveMoni(value string) error {
	data := MoniData{Moni: value}
	var err error
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	filePath := filepath.Join(cfgDir, "grandmodoro/moni_data.json")
	err = os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return err
	}
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Println(filePath)

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}

func (a *App) LoadMoni() string {
	data := MoniData{Moni: "0.0"}
	cfgDir, _ := os.UserConfigDir()
	filePath := filepath.Join(cfgDir, "grandmodoro/moni_data.json")
	file, err := os.ReadFile(filePath)
	if err == nil {
		json.Unmarshal(file, &data)
	}
	return data.Moni
}

func (a *App) IncrementMoni(value string, amount float32) (string, error) {
	numValue, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return "", err
	}
	var increment float32
	if amount >= 0 {
		increment = 0.05 + rand.Float32()*(amount-0.05)
	} else {
		increment = amount
	}
	newValue := float32(numValue) + increment
	return fmt.Sprintf("%.2f", newValue), nil
}
