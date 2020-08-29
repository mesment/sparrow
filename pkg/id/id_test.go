package id

import (
	"errors"
	"net"
	"testing"
)

func TestSonyflakeIDGenerator_Next(t *testing.T) {
	sfFactory := NewSonyflakeFactory()
	idCreator, err  := sfFactory.Create()
	if err != nil {
		t.Fatalf("create id generator error:%s", err.Error())
	}
	t.Logf("create id generator success")
	id, err := idCreator.NextID()
	if err != nil {
		t.Fatalf("create id error:%s", err.Error())
	}
	t.Logf("id:%d \n", id)


	idGenerator, err  := sfFactory.CreateWithMachineID(machineID)
	if err != nil {
		t.Fatalf("create with machineid error:%s", err.Error())
	}
	t.Logf("create with machine id success")
	id, err = idGenerator.NextID()
	if err != nil {
		t.Fatalf("create id error:%s", err.Error())
	}
	t.Logf("next id:%d \n", id)
}

func TestSonyflakeFactory_CreateWithMachineID(t *testing.T) {
	factory := NewSonyflakeFactory()

	creater, err := factory.CreateWithMachineID(machineID)
	if err != nil {
		t.Fatalf("create id generator error:%s", err.Error())
	}
	t.Logf("IDGenerator create with machineid success: %#v", creater)
}

func TestSonyflakeFactory_Create(t *testing.T) {
	sfFactory := NewSonyflakeFactory()
	idCreator, err  := sfFactory.Create()
	if err != nil {
		t.Fatalf("create id generator error:%s", err.Error())
	}
	t.Logf("IDGenerator create success: %#v", idCreator)
}

func TestDefaultIDGenerator(t *testing.T) {
	g, err := DefaultIDGenerator()
	if err != nil {
		t.Fatalf("get DefaultIDGenerator error:%v", err)
	}
	id ,err := g.NextID()
	if err != nil {
		t.Fatalf("default generator create  id error:%v", err)
	}
	t.Logf("default generator  next id:%d", id)
}


func machineID() (uint16, error) {
	ipStr := "192.168.1.135"
	ip := net.ParseIP(ipStr)
	if len(ip) < 4 {
		return 0, errors.New("invalid ipStr")
	}
	return uint16(ip[2])<<8 + uint16(ip[3]), nil
}
