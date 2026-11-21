package main

import (
	"context"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"patchdrop/internal/node"
	"patchdrop/internal/server"
)

// App is bound to the frontend; its exported methods are callable from JS.
type App struct {
	ctx  context.Context
	opts node.Options
	node *node.Node
	err  string
}

func NewApp(opts node.Options) *App { return &App{opts: opts} }

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	n, err := node.New(a.opts, a)
	if err == nil {
		err = n.Start(ctx)
	}
	if err != nil {
		a.err = err.Error()
		return
	}
	a.node = n
}

// ---- node.Events → frontend events ----

func (a *App) PeersChanged(p []node.PeerView)  { runtime.EventsEmit(a.ctx, "peers", p) }
func (a *App) TransferUpdated(t node.Transfer) { runtime.EventsEmit(a.ctx, "transfer", t) }
func (a *App) OfferClosed(id string)           { runtime.EventsEmit(a.ctx, "offerClosed", id) }
func (a *App) IncomingOffer(o server.IncomingOffer) {
	runtime.WindowUnminimise(a.ctx)
	runtime.WindowShow(a.ctx)
	runtime.EventsEmit(a.ctx, "offer", o)
}

// ---- bound methods ----

type State struct {
	Error     string          `json:"error"`
	Self      node.SelfInfo   `json:"self"`
	Peers     []node.PeerView `json:"peers"`
	Transfers []node.Transfer `json:"transfers"`
}

// GetState returns everything the UI needs on load.
func (a *App) GetState() State {
	if a.node == nil {
		return State{Error: a.err}
	}
	return State{Self: a.node.Self(), Peers: a.node.Peers(), Transfers: a.node.Transfers()}
}

func (a *App) Refresh() {
	if a.node != nil {
		a.node.Refresh()
	}
}

func (a *App) AddByAddress(addr string) (node.PeerView, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 8*time.Second)
	defer cancel()
	return a.node.AddByAddress(ctx, addr)
}

func (a *App) Scan() (int, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 45*time.Second)
	defer cancel()
	return a.node.Scan(ctx)
}

func (a *App) SetFavorite(peerID string, fav, autoAccept bool) error {
	return a.node.SetFavorite(peerID, fav, autoAccept)
}

func (a *App) PickFiles() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{Title: "Choose files to send"})
}

func (a *App) PickFolder() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Choose a folder to send"})
}

func (a *App) Send(peerID string, paths []string) (string, error) {
	return a.node.Send(peerID, paths)
}

func (a *App) CancelTransfer(id string) { a.node.CancelTransfer(id) }
func (a *App) ClearFinished()           { a.node.ClearFinished() }
func (a *App) RespondOffer(id string, accept bool) {
	a.node.RespondOffer(id, accept)
}

func (a *App) SetDeviceName(name string) error { return a.node.SetDeviceName(name) }

func (a *App) ChooseDownloadDir() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Save received files to…", DefaultDirectory: a.node.Self().DownloadDir, CanCreateDirectories: true,
	})
	if err != nil || dir == "" {
		return "", err
	}
	return dir, a.node.SetDownloadDir(dir)
}

// Reveal opens a file's folder (or the download folder when path is empty) in Finder / Explorer.
func (a *App) Reveal(path string) error {
	if path == "" {
		dir := a.node.Self().DownloadDir
		if err := a.node.SetDownloadDir(dir); err != nil { // ensures it exists
			return err
		}
		return openPath(dir, false)
	}
	return openPath(path, true)
}

func openPath(p string, selectFile bool) error {
	switch goruntime.GOOS {
	case "darwin":
		if selectFile {
			return exec.Command("open", "-R", p).Start()
		}
		return exec.Command("open", p).Start()
	case "windows":
		if selectFile {
			return exec.Command("explorer", "/select,", filepath.Clean(p)).Start()
		}
		return exec.Command("explorer", filepath.Clean(p)).Start()
	default:
		if selectFile {
			p = filepath.Dir(p)
		}
		return exec.Command("xdg-open", p).Start()
	}
}
