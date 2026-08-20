package data

import (
	"os"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
)

type Data struct {
	lock    sync.RWMutex
	path    string
	content map[string]*Content // key: channel id
}

type Content struct {
	Category []Category        `bson:"category"`
	Events   map[string]string `bson:"events"` // Key: id, Value: Title
}

type Category struct {
	Emoji string `bson:"emoji"`
	Name  string `bson:"name"`
}

func New(path string) *Data {
	return &Data{
		path:    path,
		content: make(map[string]*Content),
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
		Events:   make(map[string]string),
		Category: make([]Category, 0),
	}
	d.content[channelId] = c
	f(c)
	return false
}

func (d *Data) Sync() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.save()
}
