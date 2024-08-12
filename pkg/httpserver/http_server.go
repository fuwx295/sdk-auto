package httpserver

import (
	"sdk-auto/pkg/config"
	"sdk-auto/pkg/inspectors"
	"sdk-auto/pkg/instrument"
	"sdk-auto/pkg/process"
	"strconv"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/x/errors"
)

var instrumentor *instrument.SdkInstrumentors

func StartHttpServer(cfg *config.AutoConfig) {
	instrumentor = instrument.NewSdkInstrumentors(cfg)
	go instrumentor.Run()

	app := iris.Default()

	app.Get("/proc/list", listProc)
	app.Get("/instrument/start/{pid:int}", startInstrument)
	app.Get("/instrument/stop/{pid:int}", stopInstrument)

	err := app.Listen(":" + strconv.Itoa(cfg.Port))
	if err != nil {
		panic(err)
	}
}

func listProc(ctx iris.Context) {
	processList, err := process.FindAllProcesses(nil)
	if err != nil {
		responseWithError(ctx, err)
	}

	blackList := make([]InstrumentProcess, 0)
	instrumentList := make([]InstrumentProcess, 0)
	for _, details := range processList {
		language := inspectors.DetectLanguage(details)
		if language != nil {
			if instrumentor.IsBlackList(language, details) {
				blackList = append(blackList, NewInstrumentProcess(*language, details))
			} else {
				instrumentList = append(instrumentList, NewInstrumentProcess(*language, details))
			}
		}
	}
	ctx.JSON(iris.Map{
		"success":    true,
		"black":      blackList,
		"instrument": instrumentList,
	})
}

func startInstrument(ctx iris.Context) {
	pid := ctx.Params().GetIntDefault("pid", 0)
	if pid == 0 {
		responseWithError(ctx, errors.New("Miss Pid"))
		return
	}
	details := process.GetPidDetails(pid)
	language := inspectors.DetectLanguage(details)

	ctx.JSON(iris.Map{
		"success": true,
		"data":    instrumentor.Start(language, details),
	})
}

func stopInstrument(ctx iris.Context) {
	pid := ctx.Params().GetIntDefault("pid", 0)
	if pid == 0 {
		responseWithError(ctx, errors.New("Miss Pid"))
		return
	}

	ctx.JSON(iris.Map{
		"success": true,
		"data":    instrumentor.Stop(pid),
	})
}

func responseWithError(ctx iris.Context, err error) {
	ctx.StopWithStatus(iris.StatusInternalServerError)
	ctx.JSON(iris.Map{
		"success":  false,
		"errorMsg": err.Error(),
	})
}
