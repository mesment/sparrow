package id


type IDGenerator interface {
	NextID() (uint64, error)
}


type IDGeneratorFactory interface {
	Create() (IDGenerator, error)
	CreateWithMachineID(mid MachineID) (IDGenerator, error)
}


type MachineID    func() (uint16, error)
