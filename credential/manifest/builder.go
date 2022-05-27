package manifest

import (
	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/pkg/errors"
	"reflect"
)

const (
	BuilderEmptyError string = "builder cannot be empty"
)

type CredentialManifestBuilder struct {
	*CredentialManifest
}

func NewCredentialManifestBuilder() CredentialManifestBuilder {
	return CredentialManifestBuilder{}
}

func (cmb *CredentialManifestBuilder) Build() (*CredentialManifest, error) {
	if cmb.IsEmpty() {
		return nil, errors.New(BuilderEmptyError)
	}

	if err := cmb.CredentialManifest.IsValid(); err != nil {
		return nil, util.LoggingErrorMsg(err, "credential manifest not ready to be built")
	}

	return cmb.CredentialManifest, nil
}

func (cmb *CredentialManifestBuilder) IsEmpty() bool {
	if cmb == nil || cmb.CredentialManifest.IsEmpty() {
		return true
	}
	return reflect.DeepEqual(cmb, &CredentialManifestBuilder{})
}

type CredentialApplicationBuilder struct {
	*CredentialApplication
}

func NewCredentialApplicationBuilder() CredentialApplicationBuilder {
	return CredentialApplicationBuilder{}
}

func (cab *CredentialApplicationBuilder) Build() (*CredentialApplication, error) {
	if cab.IsEmpty() {
		return nil, errors.New(BuilderEmptyError)
	}

	if err := cab.CredentialApplication.IsValid(); err != nil {
		return nil, util.LoggingErrorMsg(err, "credential application not ready to be built")
	}

	return cab.CredentialApplication, nil
}

func (cab *CredentialApplicationBuilder) IsEmpty() bool {
	if cab == nil || cab.CredentialApplication.IsEmpty() {
		return true
	}
	return reflect.DeepEqual(cab, &CredentialApplicationBuilder{})
}

type CredentialFulfillmentBuilder struct {
	*CredentialFulfillment
}

func NewCredentialFulfillmentBuilder() CredentialFulfillmentBuilder {
	return CredentialFulfillmentBuilder{}
}

func (cfb *CredentialFulfillmentBuilder) Build() (*CredentialFulfillment, error) {
	if cfb.IsEmpty() {
		return nil, errors.New(BuilderEmptyError)
	}

	if err := cfb.CredentialFulfillment.IsValid(); err != nil {
		return nil, util.LoggingErrorMsg(err, "credential fulfillment not ready to be built")
	}

	return cfb.CredentialFulfillment, nil
}

func (cfb *CredentialFulfillmentBuilder) IsEmpty() bool {
	if cfb == nil || cfb.CredentialFulfillment.IsEmpty() {
		return true
	}
	return reflect.DeepEqual(cfb, &CredentialFulfillmentBuilder{})
}
