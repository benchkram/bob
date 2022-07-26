package bobsync

import (
	"encoding/json"
	"fmt"
	"github.com/benchkram/bob/pkg/boblog"
	"github.com/benchkram/bob/pkg/file"
	"github.com/benchkram/bob/pkg/filehash"
	"github.com/benchkram/errz"
	"github.com/logrusorgru/aurora"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

type Fingerprint struct {
	IsDir     bool
	Hash      string
	CreatedAt time.Time
}

// HashCache is a map from a file path string to a fingerprint
// paths are relative paths inside a collection to get the abs path do join (bobDir, collectionPath, thisPath)
type HashCache map[string]Fingerprint

func FromFileOrNew(path string) (hc *HashCache, err error) {
	defer errz.Recover(&err)
	fileInfo, err := os.Stat(path)
	if err != nil {
		hc := &HashCache{}
		return hc, nil
	}
	if fileInfo.IsDir() {
		return nil, fmt.Errorf("failed to load hashcache from: %s: it is a directory", path)
	}
	f, err := os.Open(path)
	errz.Fatal(err)
	defer f.Close()
	byteValue, err := ioutil.ReadAll(f)
	errz.Fatal(err)
	err = json.Unmarshal(byteValue, &hc)
	errz.Fatal(err)
	return hc, nil
}

func (h *HashCache) SaveToFile(path string) (err error) {
	defer errz.Recover(&err)
	data, err := json.Marshal(*h)
	errz.Fatal(err)
	if file.Exists(path) {
		err := os.Remove(path)
		errz.Fatal(err)
	}
	err = ioutil.WriteFile(path, data, 0644)
	errz.Fatal(err)
	return nil
}

func (h *HashCache) Update(basePath string) (err error) {
	defer errz.Recover(&err)
	var filePaths []string
	err = filepath.Walk(basePath,
		func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == basePath {
				return nil
			}
			filePaths = append(filePaths, path)
			return nil
		})
	errz.Fatal(err)

	saveMap := h.toInternalMap(filePaths, basePath)

	// analogue to https://gobyexample.com/worker-pools
	numJobs := len(filePaths)
	jobs := make(chan string, numJobs)
	results := make(chan error, numJobs)

	numWorkers := runtime.NumCPU()
	for w := 1; w <= numWorkers; w++ {
		go updater(saveMap, jobs, results)
	}
	for _, f := range filePaths {
		jobs <- f
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		e := <-results
		if e != nil {
			boblog.Log.Error(e, e.Error())
			err = e
		}
	}
	if err != nil {
		fmt.Println(aurora.Red("failed to hash some files and will ignore them"))
	}
	err = h.overrideFromInternalMap(saveMap, basePath)
	errz.Fatal(err)
	return nil

}

func updater(saveMap *sync.Map, files <-chan string, result chan<- error) {
	for f := range files {
		var reHash bool
		oldFingerprint, ok := saveMap.Load(f)
		if ok {
			var oldFingerprint = oldFingerprint.(Fingerprint)
			lastMod, err := file.LastModTime(f)
			if err != nil {
				result <- err
				continue
			}
			if oldFingerprint.CreatedAt.Before(lastMod) {
				reHash = true
			}
		} else {
			reHash = true
		}
		if reHash {
			fi, err := os.Stat(f)
			if err != nil {
				result <- err
				continue
			}
			var h string
			isDir := fi.IsDir()
			if isDir {
				hashBytes, err := filehash.HashString(filepath.Base(f))
				if err != nil {
					result <- err
					continue
				}
				h = filehash.HashToString(hashBytes)
			} else {
				hashBytes, err := filehash.Hash(f)
				h = filehash.HashToString(hashBytes)
				if err != nil {
					result <- err
					continue
				}
			}
			fp := Fingerprint{
				Hash:      h,
				CreatedAt: time.Now(),
				IsDir:     isDir,
			}
			saveMap.Store(f, fp)
		}
		result <- nil
	}
}

// toInternalMap copies all paths in the filePaths argument from the HashCache to a sync.Map which can be safely used in
//concurrent runs
func (h *HashCache) toInternalMap(filePaths []string, basePath string) *sync.Map {
	saveMap := &sync.Map{}
	for _, f := range filePaths {
		fp, ok := (*h)[f]
		if ok {
			saveMap.Store(filepath.Join(basePath, f), fp)
		}
	}
	return saveMap
}

func (h *HashCache) overrideFromInternalMap(saveMap *sync.Map, basePath string) (err error) {
	defer errz.Recover(&err)
	for k := range *h {
		delete(*h, k)
	}
	saveMap.Range(func(key, value interface{}) bool {
		relPath, err := filepath.Rel(basePath, key.(string))
		errz.Fatal(err)
		(*h)[relPath] = value.(Fingerprint)
		return true
	})
	return nil
}

func (h *HashCache) SortedKeys() []string {
	keys := make([]string, 0)
	for k := range *h {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
