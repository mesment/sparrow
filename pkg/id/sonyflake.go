package id

import (
	"errors"
	"github.com/sony/sonyflake"
)


type SonyflakeFactory struct {
}

func NewSonyflakeFactory() *SonyflakeFactory  {
	return &SonyflakeFactory{}
}

func (SonyflakeFactory) Create() (IDGenerator, error) {

	st := sonyflake.Settings{}

	sf := sonyflake.NewSonyflake(st)

	if sf == nil {
		return nil, errors.New("failed to create sonyflake")
	}
	return &sonyflakeIDGenerator{sf:sf}, nil
}

func (SonyflakeFactory) CreateWithMachineID(mid MachineID) (IDGenerator, error ){
	st := sonyflake.Settings{}

	if mid != nil {
		st.MachineID = mid
	}
	sf := sonyflake.NewSonyflake(st)
	if sf == nil {
		return nil, errors.New("failed to create sonyflake")
	}

	return &sonyflakeIDGenerator{sf:sf}, nil
}


type sonyflakeIDGenerator struct {
	sf *sonyflake.Sonyflake
}

func (s sonyflakeIDGenerator) NextID()(uint64, error) {
	return s.sf.NextID()
}


