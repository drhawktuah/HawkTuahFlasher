package hawktuah

type ValueType uint8

type Definition struct {
	Name   				string
	Vendor 				string
	Family 				string

	NameDocumentation   Documentation
	VendorDocumentation Documentation
	FamilyDocumentation Documentation

	Documentation 		Documentation

	Detect   			Detection
	Protocol 			Protocol
	Flash    			Flash

	Custom map[string]  Value
	CMake  map[string]  Value
}

type Value struct {
	Type 		  ValueType

	String  	  string
	Number  	  uint64
	Boolean 	  bool

	Documentation PropertyDocumentation
}

type Detection struct {
	VIDs []DetectionVID
}

type DetectionVID struct {
	Value         uint16
	Documentation PropertyDocumentation
}

type Protocol struct {
	Bootloader              string
	BootloaderDocumentation PropertyDocumentation
}

type Flash struct {
	Baudrate              uint32
	BaudrateDocumentation PropertyDocumentation

	Erase                 bool
	EraseDocumentation    PropertyDocumentation

	Verify                bool
	VerifyDocumentation   PropertyDocumentation
}

type Documentation struct {
	Tags map[string]string
}

type PropertyDocumentation struct {
	Tags map[string]string
}

func NewDefinition() *Definition {
	return &Definition{
		Documentation: NewDocumentation(),

		Custom: make(map[string]Value),
		CMake:  make(map[string]Value),
	}
}

func NewDocumentation() Documentation {
	return Documentation{
		Tags: make(map[string]string),
	}
}

func NewPropertyDocumentation() PropertyDocumentation {
	return PropertyDocumentation{
		Tags: make(map[string]string),
	}
}