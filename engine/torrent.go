package engine

import (
	"time"
	"github.com/anacrolix/torrent"
)

type Torrent struct {
	//anacrolix/torrent
	InfoHash   string 				`json:"infoHash"`
	Name       string				`json:"name"`
	Loaded     bool					`json:"loaded"`
	Downloaded int64				`json:"downloaded"`
	Size       int64				`json:"size"`
	Files      []*File				`json:"files"`
	//cloud torrent
	Started      bool				`json:"started"`
	Dropped      bool				`json:"dropped"`
	Percent      float32			`json:"percent"`
	DownloadRate float32			`json:"downloadRate"`
	t            *torrent.Torrent
	updatedAt    time.Time			`json:"updatedAt"`
}

type File struct {
	//anacrolix/torrent
	Path      string
	Size      int64
	Chunks    int
	Completed int
	//cloud torrent
	Started bool
	Percent float32
	f       torrent.File
}

func (torrent *Torrent) Update(t *torrent.Torrent) {
	torrent.Name = t.Name()
	torrent.Loaded = t.Info() != nil
	if torrent.Loaded {
		torrent.updateLoaded(t)
	}
	torrent.t = t
}

func (torrent *Torrent) updateLoaded(t *torrent.Torrent) {

	torrent.Size = t.Length()
	totalChunks := 0
	totalCompleted := 0

	tfiles := t.Files()
	if len(tfiles) > 0 && torrent.Files == nil {
		torrent.Files = make([]*File, len(tfiles))
	}
	//merge in files
	for i, f := range tfiles {
		path := f.Path()
		file := torrent.Files[i]
		if file == nil {
			file = &File{Path: path}
			torrent.Files[i] = file
		}
		chunks := f.State()

		file.Size = f.Length()
		file.Chunks = len(chunks)
		completed := 0
		for _, p := range chunks {
			if p.Complete {
				completed++
			}
		}
		file.Completed = completed
		file.Percent = percent(int64(file.Completed), int64(file.Chunks))
		file.f = f

		totalChunks += file.Chunks
		totalCompleted += file.Completed
	}

	//cacluate rate
	now := time.Now()
	bytes := t.BytesCompleted()
	torrent.Percent = percent(bytes, torrent.Size)
	if !torrent.updatedAt.IsZero() {
		dt := float32(now.Sub(torrent.updatedAt))
		db := float32(bytes - torrent.Downloaded)
		rate := db * (float32(time.Second) / dt)
		if rate >= 0 {
			torrent.DownloadRate = rate
		}
	}
	torrent.Downloaded = bytes
	torrent.updatedAt = now
}

func percent(n, total int64) float32 {
	if total == 0 {
		return float32(0)
	}
	return float32(int(float64(10000)*(float64(n)/float64(total)))) / 100
}
