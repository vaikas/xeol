package ui

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	syftUI "github.com/anchore/syft/ui"
	"github.com/dustin/go-humanize"
	"github.com/gookit/color"
	"github.com/wagoodman/go-partybus"
	"github.com/wagoodman/go-progress"
	"github.com/wagoodman/go-progress/format"
	"github.com/wagoodman/jotframe/pkg/frame"

	"github.com/xeol-io/xeol/internal/ui/components"
	xeolEventParsers "github.com/xeol-io/xeol/xeol/event/parsers"
)

const maxBarWidth = 50
const statusSet = components.SpinnerDotSet // SpinnerCircleOutlineSet
const completedStatus = "✔"                // "●"
const tileFormat = color.Bold

var auxInfoFormat = color.HEX("#777777")
var statusTitleTemplate = fmt.Sprintf(" %%s %%-%ds ", syftUI.StatusTitleColumn)

func startProcess() (format.Simple, *components.Spinner) {
	width, _ := frame.GetTerminalSize()
	barWidth := int(0.25 * float64(width))
	if barWidth > maxBarWidth {
		barWidth = maxBarWidth
	}
	formatter := format.NewSimpleWithTheme(barWidth, format.HeavyNoBarTheme, format.ColorCompleted, format.ColorTodo)
	spinner := components.NewSpinner(statusSet)

	return formatter, &spinner
}

func (r *Handler) EolScanningStartedHandler(ctx context.Context, fr *frame.Frame, event partybus.Event, wg *sync.WaitGroup) error {
	monitor, err := xeolEventParsers.ParseEolScanningStarted(event)
	if err != nil {
		return fmt.Errorf("bad %s event: %w", event.Type, err)
	}

	line, err := fr.Append()
	if err != nil {
		return err
	}

	wg.Add(1)

	_, spinner := startProcess()
	stream := progress.StreamMonitors(ctx, []progress.Monitorable{monitor.PackagesProcessed, monitor.EolDiscovered}, 50*time.Millisecond)
	title := tileFormat.Sprint("Scanning image...")

	formatFn := func(val int64) {
		spin := color.Magenta.Sprint(spinner.Next())
		auxInfo := auxInfoFormat.Sprintf("[%d eol packages]", val)
		_, _ = io.WriteString(line, fmt.Sprintf(statusTitleTemplate+"%s", spin, title, auxInfo))
	}

	go func() {
		defer wg.Done()

		formatFn(0)
		for p := range stream {
			formatFn(p[1])
		}

		spin := color.Green.Sprint(completedStatus)
		title = tileFormat.Sprint("Scanned image")
		auxInfo := auxInfoFormat.Sprintf("[%d eol]", monitor.EolDiscovered.Current())
		_, _ = io.WriteString(line, fmt.Sprintf(statusTitleTemplate+"%s", spin, title, auxInfo))
	}()

	return nil
}

func (r *Handler) UpdateEolDatabaseHandler(ctx context.Context, fr *frame.Frame, event partybus.Event, wg *sync.WaitGroup) error {
	prog, err := xeolEventParsers.ParseUpdateEolDatabase(event)
	if err != nil {
		return fmt.Errorf("bad FetchImage event: %w", err)
	}

	line, err := fr.Prepend()
	if err != nil {
		return err
	}

	wg.Add(1)

	formatter, spinner := startProcess()
	stream := progress.Stream(ctx, prog, 150*time.Millisecond)
	title := tileFormat.Sprint("EOL DB")

	formatFn := func(p progress.Progress) {
		progStr, err := formatter.Format(p)
		spin := color.Magenta.Sprint(spinner.Next())
		if err != nil {
			_, _ = io.WriteString(line, fmt.Sprintf("Error: %+v", err))
		} else {
			var auxInfo string
			switch prog.Stage() {
			case "downloading":
				progStr += " "
				auxInfo = auxInfoFormat.Sprintf(" [%s / %s]", humanize.Bytes(uint64(prog.Current())), humanize.Bytes(uint64(prog.Size())))
			default:
				progStr = ""
				auxInfo = auxInfoFormat.Sprintf("[%s]", prog.Stage())
			}

			_, _ = io.WriteString(line, fmt.Sprintf(statusTitleTemplate+"%s%s", spin, title, progStr, auxInfo))
		}
	}

	go func() {
		defer wg.Done()

		formatFn(progress.Progress{})
		for p := range stream {
			formatFn(p)
		}

		spin := color.Green.Sprint(completedStatus)
		title = tileFormat.Sprint("EOL DB")
		auxInfo := auxInfoFormat.Sprintf("[%s]", prog.Stage())
		_, _ = io.WriteString(line, fmt.Sprintf(statusTitleTemplate+"%s", spin, title, auxInfo))
	}()
	return err
}
