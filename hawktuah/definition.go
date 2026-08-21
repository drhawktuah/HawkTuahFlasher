package hawktuah

type ValueType uint8

type Definition struct {
	Name   string
	Vendor string
	Family string

	NameDocumentation   Documentation
	VendorDocumentation Documentation
	FamilyDocumentation Documentation

	Documentation Documentation

	Detect   Detection
	Protocol Protocol
	Flash    Flash

	Custom map[string]Value
	CMake  map[string]Value
}

type Value struct {
	Type ValueType

	String  string
	Number  uint64
	Boolean bool

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

	Erase              bool
	EraseDocumentation PropertyDocumentation

	Verify              bool
	VerifyDocumentation PropertyDocumentation
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

func (definition *Definition) IsValid() bool {
	return definition != nil && definition.Validate() == nil
}

func (definition *Definition) Validate() error {
	return ValidateDefinition(definition)
}

func (value Value) AsString() (string, bool) {
	if value.Type != ValueString {
		return "", false
	}

	return value.String, true
}

func (value Value) AsNumber() (uint64, bool) {
	if value.Type != ValueNumber {
		return 0, false
	}

	return value.Number, true
}

func (value Value) AsBoolean() (bool, bool) {
	if value.Type != ValueBoolean {
		return false, false
	}

	return value.Boolean, true
}

func (definition *Definition) FindVID(vid uint16) *DetectionVID {
	if definition == nil {
		return nil
	}

	for index := range definition.Detect.VIDs {
		detection := &definition.Detect.VIDs[index]

		if detection.Value == vid {
			return detection
		}
	}

	return nil
}