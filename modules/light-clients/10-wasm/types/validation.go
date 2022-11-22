package types

import cosmwasm "github.com/CosmWasm/wasmvm"

// Basic validation config can be extended to add other configuration later
type ValidationConfig struct {
	MaxSizeAllowed int
}

func NewWasmValidator(config *ValidationConfig, vmCreateFn func() (*cosmwasm.VM, error)) (*WasmValidator, error) {
	return &WasmValidator{
		config:     config,
		vmCreateFn: vmCreateFn,
	}, nil
}

type WasmValidator struct {
	vmCreateFn func() (*cosmwasm.VM, error)
	config     *ValidationConfig
}

func (v *WasmValidator) validateWasmCode(code []byte) (bool, error) {
	if len(code) > v.config.MaxSizeAllowed {
		return false, nil
	}

	testVM, err := v.vmCreateFn()
	if err != nil {
		return false, err
	}

	_, err = testVM.Create(code)
	if err != nil {
		return false, nil
	}

	// TODO: validate the cosmwasm code interface to ensure that it exposed the correct entry points
	testVM.Cleanup()
	return true, nil
}