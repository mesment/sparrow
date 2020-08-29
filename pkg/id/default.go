package id

import "sync"

var defaultIDGenerator IDGenerator
var lock sync.Mutex

func DefaultIDGenerator() (IDGenerator, error)  {
	if defaultIDGenerator == nil {
		lock.Lock()
		defer lock.Unlock()
		if defaultIDGenerator == nil {
			f := NewSonyflakeFactory()
			defaultIDGenerator, err := f.Create()
			if err != nil {
				return nil ,err
			}
			return defaultIDGenerator, nil
		}
	}
	return defaultIDGenerator, nil
}
