package hyaml

import (
	"os"
	"sync"
)

func Init[T any](path string) *Config[T] {
	var c Config[T]
	c.path = path
	return &c
}

func Load[T any](path string, t *T) *Config[T] {
	return &Config[T]{
		path:   path,
		Config: t,
	}
}

func LoadFormPath[T any](path string) (*Config[T], error) {
	var bc Config[T]
	err := bc.Load(path)
	if err != nil {
		return nil, err
	}
	return &bc, nil
}

func LoadFormBytes[T any](b []byte) (*Config[T], error) {
	var bc Config[T]
	err := bc.LoadFromBytes(b)
	if err != nil {
		return nil, err
	}
	return &bc, nil
}

type Config[T any] struct {
	mux    sync.Mutex
	path   string
	Config *T
}

func (bc *Config[T]) Update() error {
	if bc.path == "" {
		return nil
	}
	return bc.Load(bc.path)
}

func (bc *Config[T]) Load(path string) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	bc.path = path
	var t T
	err = Unmarshal(b, &t)
	if err != nil {
		return err
	}
	bc.Config = &t
	return nil
}

func (bc *Config[T]) LoadFromBytes(b []byte) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	var t T
	err := Unmarshal(b, &t)
	if err != nil {
		return err
	}
	bc.Config = &t
	return nil
}

func (bc *Config[T]) Save() error {
	if bc.path == "" {
		return nil
	}
	return bc.SaveAs(bc.path)
}

func (bc *Config[T]) SaveAs(path string) error {
	bc.mux.Lock()
	defer bc.mux.Unlock()
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	b, err := MarshalWithComment(bc.Config)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func (bc *Config[T]) SetPath(path string) {
	bc.path = path
}

func SavePath(path string, a any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	b, err := MarshalWithComment(a)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	if err != nil {
		return err
	}
	return nil
}
