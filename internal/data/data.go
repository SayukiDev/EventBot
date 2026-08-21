package data

import (
	"EventBot/log"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.uber.org/zap"
)

type Data struct {
	lock    sync.RWMutex
	path    string
	content map[string]*Content // key: channel id
	close   chan struct{}
	once    sync.Once
}

type Content struct {
	Category     []Category `bson:"category"`
	DefaultEvent string     `bson:"default_event"`
}

type Category struct {
	Emoji string `bson:"emoji"`
	Name  string `bson:"name"`
}

func New(path string) *Data {
	return &Data{
		path:    path,
		content: make(map[string]*Content),
		close:   make(chan struct{}, 1),
	}
}

func (d *Data) Load() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	bytes, err := os.ReadFile(d.path)
	if err != nil {
		if os.IsNotExist(err) {
			return d.save()
		}
		return err
	}
	err = bson.Unmarshal(bytes, d.content)
	if err != nil {
		return err
	}
	return nil
}

func (d *Data) save() error {
	bytes, err := bson.Marshal(d.content)
	if err != nil {
		return err
	}
	return os.WriteFile(d.path, bytes, 0644)
}

func (d *Data) Get(channelId string, f func(c *Content)) bool {
	d.lock.RLock()
	defer d.lock.RUnlock()
	if c, ok := d.content[channelId]; ok {
		f(c)
		return true
	} else {
		return false
	}
}

func (d *Data) Set(channelId string, f func(c *Content)) bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	if c, ok := d.content[channelId]; ok {
		f(c)
		return true
	}

	c := &Content{
		Category: make([]Category, 0),
	}
	d.content[channelId] = c
	f(c)
	return false
}

func (d *Data) Start() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	go func() {
		for {
			select {
			case <-d.close:
				close(d.close)
				return
			default:
			}
			func() {
				d.lock.Lock()
				defer d.lock.Unlock()
				derr := d.save()
				if derr != nil {
					log.ErrorE("Failed to save data", zap.Error(derr))
				}
			}()
			time.Sleep(time.Second * 360)
		}
	}()
	return nil
}

func (d *Data) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	var err error
	d.once.Do(func() {
		d.close <- struct{}{}
		err = d.save()
	})
	return err
}
