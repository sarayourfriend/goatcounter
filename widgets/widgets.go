// Copyright © 2019 Martin Tournoij – This file is part of GoatCounter and
// published under the terms of a slightly modified EUPL v1.2 license, which can
// be found in the LICENSE file or at https://license.goatcounter.com

package widgets

import (
	"context"
	"fmt"
	"html/template"
	"time"

	"zgo.at/goatcounter"
	"zgo.at/zstd/zint"
)

type (
	Widget interface {
		GetData(context.Context, Args) error
		RenderHTML(context.Context, SharedData) (string, interface{})

		SetHTML(template.HTML)
		HTML() template.HTML
		SetErr(error)
		Err() error

		Name() string
		Type() string // "full-width", "hchart"
		Label() string
	}

	Args struct {
		Start, End  time.Time
		PathFilter  []int64
		Daily       bool
		ForcedDaily bool
		ShowRefs    string
		AsText      bool
	}

	// SharedData gets passed to every widget.
	SharedData struct {
		Site *goatcounter.Site
		Args Args

		Total             int
		TotalUnique       int
		TotalUniqueUTC    int
		TotalEvents       int
		TotalEventsUnique int
	}
)

type List []Widget

var (
	ShowRefs       zint.Bitflag8 = 0b0001
	FilterInternal zint.Bitflag8 = 0b0010
	FilterOff      zint.Bitflag8 = 0b0100
)

func FromSiteWidgets(www goatcounter.Widgets, params zint.Bitflag8) List {
	widgetList := make(List, 0, len(www)+4)
	if !params.Has(FilterInternal) {
		// We always need these to know the total number of pageviews.
		widgetList = append(widgetList, NewWidget("totalcount"))
	}
	for _, w := range www {
		if params.Has(FilterOff) && !w["on"].(bool) {
			continue
		}

		name := w["name"].(string)
		ww := NewWidget(name)

		switch name {
		case "pages":
			if !params.Has(FilterInternal) {
				widgetList = append(widgetList, NewWidget("max"))
			}
			if params.Has(ShowRefs) {
				widgetList = append(widgetList, NewWidget("refs"))
			}

			wp := ww.(*Pages)
			if n, ok := w["limit_pages"].(float64); ok {
				wp.LimitPage = int(n)
			}
			if n, ok := w["limit_ref"].(float64); ok {
				wp.LimitRef = int(n)
			}
			ww = wp
		}
		widgetList = append(widgetList, ww)
	}

	return widgetList
}

// Get a widget from the list by name.
func (l List) Get(name string) Widget {
	for _, w := range l {
		if w.Name() == name {
			return w
		}
	}
	return nil
}

func NewWidget(name string) Widget {
	switch name {
	case "totalcount":
		return &TotalCount{}
	case "max":
		return &Max{}
	case "refs":
		return &Refs{}

	case "pages":
		return &Pages{}
	case "totalpages":
		return &TotalPages{}
	case "toprefs":
		return &TopRefs{}
	case "browsers":
		return &Browsers{}
	case "systems":
		return &Systems{}
	case "sizes":
		return &Sizes{}
	case "locations":
		return &Locations{}
	}
	panic(fmt.Errorf("unknown widget: %q", name))
}

func (w *TotalCount) GetData(ctx context.Context, a Args) (err error) {
	w.Total, w.TotalUnique, w.TotalUniqueUTC, w.TotalEvents, w.TotalEventsUnique, err = goatcounter.GetTotalCount(ctx, a.Start, a.End, a.PathFilter)
	return err
}

func (w *Pages) GetData(ctx context.Context, a Args) (err error) {
	w.Display, w.UniqueDisplay, w.More, err = w.Pages.List(
		ctx, a.Start, a.End, a.PathFilter, nil, a.Daily)
	return err
}
func (w *Max) GetData(ctx context.Context, a Args) (err error) {
	w.Max, err = goatcounter.GetMax(ctx, a.Start, a.End, a.PathFilter, a.Daily)
	return err
}
func (w *TotalPages) GetData(ctx context.Context, a Args) (err error) {
	w.Max, err = w.Total.Totals(ctx, a.Start, a.End, a.PathFilter, a.Daily)
	return err
}
func (w *Refs) GetData(ctx context.Context, a Args) (err error) {
	return w.Refs.ListRefsByPath(ctx, a.ShowRefs, a.Start, a.End, 0)
}
func (w *TopRefs) GetData(ctx context.Context, a Args) (err error) {
	return w.TopRefs.ListTopRefs(ctx, a.Start, a.End, a.PathFilter, 0)
}
func (w *Browsers) GetData(ctx context.Context, a Args) (err error) {
	return w.Browsers.ListBrowsers(ctx, a.Start, a.End, a.PathFilter, 6, 0)
}
func (w *Systems) GetData(ctx context.Context, a Args) (err error) {
	return w.Systems.ListSystems(ctx, a.Start, a.End, a.PathFilter, 6, 0)
}
func (w *Sizes) GetData(ctx context.Context, a Args) (err error) {
	return w.SizeStat.ListSizes(ctx, a.Start, a.End, a.PathFilter)
}
func (w *Locations) GetData(ctx context.Context, a Args) (err error) {
	return w.LocStat.ListLocations(ctx, a.Start, a.End, a.PathFilter, 6, 0)
}
