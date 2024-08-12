package instrument

import (
	"context"
	"fmt"
	"log"
	"sdk-auto/pkg/config"
	"sdk-auto/pkg/inspectors"
	"sdk-auto/pkg/process"
	"sync"
	"time"
)

type SdkFactory interface {
	NewInstrumentor(ctx context.Context, endPoint string, details *process.Details) (SdkInstrumentor, error)
}

type SdkInstrumentor interface {
	Run(ctx context.Context) error
	Close(ctx context.Context) error
}

type SdkInstrumentors struct {
	mux             sync.Mutex
	endPoint        string
	scanInterval    int64
	lastScannedPids map[int]bool
	toInstruments   map[inspectors.ProgrammingLanguage][]*process.Details
	instrumented    map[int]SdkInstrumentor
	sdks            map[inspectors.ProgrammingLanguage]SdkFactory
	blackList       map[inspectors.ProgrammingLanguage]Match
}

func NewSdkInstrumentors(cfg *config.AutoConfig) *SdkInstrumentors {
	return &SdkInstrumentors{
		instrumented:  make(map[int]SdkInstrumentor),
		endPoint:      cfg.TraceApi,
		scanInterval:  cfg.ScanInterval,
		toInstruments: make(map[inspectors.ProgrammingLanguage][]*process.Details, 0),
		sdks: map[inspectors.ProgrammingLanguage]SdkFactory{
			inspectors.GoProgrammingLanguage: &GoInstrumentorFactory{},
		},
		blackList: map[inspectors.ProgrammingLanguage]Match{
			inspectors.GoProgrammingLanguage:   NewCommMatch(cfg.BlackList.GetGoList()),
			inspectors.JavaProgrammingLanguage: NewCmdLineMatch(cfg.BlackList.GetJavaList()),
		},
	}
}

func (i *SdkInstrumentors) Run() {
	if i.scanInterval > 0 {
		timer := time.NewTicker(time.Duration(i.scanInterval) * time.Second)
		for {
			select {
			case <-timer.C:
				if len(i.toInstruments) > 0 {
					// 滞后一个周期再进行Instrument.
					for lang, details := range i.toInstruments {
						for _, detail := range details {
							i.Start(&lang, detail)
						}
					}
					// 此处无需加锁，因为只有该定时任务对该变量操作
					i.toInstruments = make(map[inspectors.ProgrammingLanguage][]*process.Details, 0)
				}
				processList, err := process.FindAllProcesses(nil)
				if err != nil {
					log.Printf("Error to find process: %v", err)
					return
				}
				currentPids := make(map[int]bool)
				for _, details := range processList {
					currentPids[details.ProcessID] = true
					if _, exist := i.lastScannedPids[details.ProcessID]; exist {
						continue
					}
					language := inspectors.DetectLanguage(details)
					if i.check(language, details) != "" {
						continue
					}
					if inspectors.GoProgrammingLanguage != *language {
						// 现暂只支持GO的自动识别，后续可调整为Java也支持
						continue
					}

					log.Printf("Add %d-%s to instrumentList", details.ProcessID, details.Comm)
					toInstrumentList, ok := i.toInstruments[*language]
					if !ok {
						toInstrumentList = make([]*process.Details, 0)
					}
					toInstrumentList = append(toInstrumentList, details)
					i.toInstruments[*language] = toInstrumentList
				}
				// 重新设置已扫描PID列表，由于只有该定时任务修改该变量，无需加锁
				i.lastScannedPids = currentPids

				// 存在进程已销毁场景，需清理进程
				for pid := range i.instrumented {
					if _, exist := i.lastScannedPids[pid]; !exist {
						i.Stop(pid)
					}
				}
			}
		}
	}
}

func (i *SdkInstrumentors) IsBlackList(language *inspectors.ProgrammingLanguage, details *process.Details) bool {
	if match, exist := i.blackList[*language]; exist {
		return match.IsMatch(details)
	}
	return false
}

func (i *SdkInstrumentors) check(language *inspectors.ProgrammingLanguage, details *process.Details) string {
	if language == nil {
		return fmt.Sprintf("Unknown Language: %s", details.Comm)
	}
	if _, exist := i.sdks[*language]; !exist {
		return fmt.Sprintf("Language %s is not supported", *language)
	}
	if match, exist := i.blackList[*language]; exist {
		if match.IsMatch(details) {
			log.Printf("Ignore BlackList Process %s, pid: %d", details.Comm, details.ProcessID)
			return fmt.Sprintf("Ignore BlackList Process %s", details.Comm)
		}
	}
	if _, exists := i.instrumented[details.ProcessID]; exists {
		return fmt.Sprintf("Process %d is already instrumented", details.ProcessID)
	}
	if process.GetComm(details.ProcessID) == "" {
		log.Printf("Process %d-%s is not exist", details.ProcessID, details.Comm)
		return fmt.Sprintf("Process %d is not exist", details.ProcessID)
	}
	return ""
}

func (i *SdkInstrumentors) Start(language *inspectors.ProgrammingLanguage, details *process.Details) string {
	if errMsg := i.check(language, details); errMsg != "" {
		return errMsg
	}

	log.Printf("Start instrument %s, Process %s, pid: %d", *language, details.Comm, details.ProcessID)

	go func() {
		sdk := i.sdks[*language]
		inst, err := sdk.NewInstrumentor(context.Background(), i.endPoint, details)
		if err != nil {
			return
		}

		i.mux.Lock()
		i.instrumented[details.ProcessID] = inst
		i.mux.Unlock()

		if err := inst.Run(context.Background()); err != nil {
			log.Printf("instrumentation crashed: %v", err)
		}
	}()
	return fmt.Sprintf("Success Start instrument %d", details.ProcessID)
}

func (i *SdkInstrumentors) Stop(pid int) string {
	log.Printf("Stop instrument %d", pid)

	if inst, exist := i.instrumented[pid]; exist {
		i.mux.Lock()
		delete(i.instrumented, pid)
		i.mux.Unlock()

		go func() {
			err := inst.Close(context.Background())
			if err != nil {
				log.Printf("Error stop instrument process %d: %v", pid, err)
			}
		}()
		return fmt.Sprintf("Success Stop instrument %d", pid)
	} else {
		return fmt.Sprintf("%d is not instrumented", pid)
	}
}
