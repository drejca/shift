package wasm

var (
	WASM_MAGIC_NUM = []byte{0x00, 0x61, 0x73, 0x6d}
	WASM_VERSION_1 = []byte{0x01, 0x00,0x00, 0x00}
)

const (
	ZERO byte = 0x00

	// Function Bodies
	BODY_END = 0x0b

	// Module sections
	SECTION_TYPE = 0x01
	SECTION_FUNC = 0x03
	SECTION_EXPORT = 0x07
	SECTION_CODE = 0x0a

	// Language Types
	TYPE_I32 = 0x7f
	FUNC = 0x60

	// external_kind kind for import/export
	EXT_KIND_FUNC = 0x00

	// Constants
	CONST_I32 = 0x41
)
