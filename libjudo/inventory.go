package libjudo

import (
	"bufio"
	"io"
	"os"
	"path"
	"regexp"
	"time"
)

type Inventory struct {
	hosts []*Host
	names []string
	s     *SeenString
}

func NewInventory(names []string) *Inventory {
	return &Inventory{
		hosts: []*Host{},
		names: names,
		s:     NewSeenString(),
	}
}

func (inventory *Inventory) Populate() {
	for _, name := range inventory.names {
		for host := range inventory.resolveNames(name) {
			inventory.hosts = append(inventory.hosts, host)
		}
	}
}

func (inventory *Inventory) GetHosts() (ch chan *Host) {
	ch = make(chan *Host)
	go func() {
		for _, host := range inventory.hosts {
			ch <- host
		}
		close(ch)
	}()
	return
}

func isExecutable(mode os.FileMode) bool {
	return (mode.Perm() & 0111) > 0
}

func readGroups(r io.Reader) (out []string) {
	scanner := bufio.NewScanner(r)
	re := regexp.MustCompile("^[^# ]+")
	for scanner.Scan() {
		line := scanner.Text()
		host := re.FindString(line)
		if host == "" {
			continue
		}
		out = append(out, host)
	}
	assert(scanner.Err())
	return
}

func (inventory *Inventory) resolveNames(name string) (ch chan *Host) {
	ch = make(chan *Host)
	fname := path.Join("groups", name)
	stat, err := os.Stat(fname)

	if err != nil {
		go func() {
			if !inventory.s.SeenBefore(name) {
				ch <- NewHost(name)
			}
			close(ch)
		}()
		return
	}

	if !stat.Mode().IsRegular() {
		close(ch)
		panic("not regular file")
	}
	go func() {
		// var f io.Reader
		defer close(ch)
		if isExecutable(stat.Mode()) {
			proc, err := NewProc(fname)
			assert(err)
			close(proc.Stdin)
			for {
				select {
				case line := <-proc.Stdout:
					for host := range inventory.resolveNames(line) {
						ch <- host
					}
				case line := <-proc.Stderr:
					logger.Print(line)
				case err = <-proc.Done:
					assert(err)
					return
				case <-time.After(10 * time.Second):
					panic(TimeoutError{})
				}
			}

		} else {
			f, err := os.Open(fname)
			assert(err)
			defer f.Close()
			for _, name_ := range readGroups(f) {
				for host := range inventory.resolveNames(name_) {
					ch <- host
				}
			}
		}

	}()
	return
}
