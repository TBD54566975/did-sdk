package manifest

import (
	"github.com/TBD54566975/ssi-sdk/credential/exchange"
	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/pkg/errors"
	"reflect"
)

// CredentialManifest https://identity.foundation/credential-manifest/#general-composition
type CredentialManifest struct {
	ID                     string                           `json:"id" validate:"required"`
	Issuer                 Issuer                           `json:"issuer" validate:"required,dive"`
	OutputDescriptors      []OutputDescriptor               `json:"output_descriptors" validate:"required,dive"`
	Format                 *exchange.ClaimFormat            `json:"format,omitempty" validate:"omitempty,dive"`
	PresentationDefinition *exchange.PresentationDefinition `json:"presentation_definition,omitempty" validate:"omitempty,dive"`
}

func (cm *CredentialManifest) IsEmpty() bool {
	if cm == nil {
		return true
	}
	return reflect.DeepEqual(cm, &CredentialManifest{})
}

func (cm *CredentialManifest) IsValid() error {
	if cm.IsEmpty() {
		return errors.New("manifest is empty")
	}
	if err := IsValidCredentialManifest(*cm); err != nil {
		return errors.Wrap(err, "manifest failed json schema validation")
	}
	if err := AreValidOutputDescriptors(cm.OutputDescriptors); err != nil {
		return errors.Wrap(err, "manifest's output descriptors failed json schema validation")
	}
	return util.NewValidator().Struct(cm)
}

type Issuer struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name,omitempty"`
	// an object or URI as defined by the DIF Entity Styles specification
	// https://identity.foundation/wallet-rendering/#entity-styles
	// TODO(gabe) https://github.com/TBD54566975/ssi-sdk/issues/52
	Styles interface{} `json:"styles,omitempty"`
}

// OutputDescriptor https://identity.foundation/credential-manifest/#output-descriptor
type OutputDescriptor struct {
	// Must be unique within a manifest
	ID          string `json:"id" validate:"required"`
	Schema      string `json:"schema" validate:"required"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// both below: an object or URI as defined by the DIF Entity Styles specification
	// https://identity.foundation/wallet-rendering/#entity-styles
	// TODO(gabe) https://github.com/TBD54566975/ssi-sdk/issues/52
	Styles  interface{} `json:"styles,omitempty"`
	Display interface{} `json:"display,omitempty"`
}

func (od *OutputDescriptor) IsEmpty() bool {
	if od == nil {
		return true
	}
	return reflect.DeepEqual(od, &OutputDescriptor{})
}

func (od *OutputDescriptor) IsValid() error {
	return util.NewValidator().Struct(od)
}

// CredentialApplication https://identity.foundation/credential-manifest/#credential-application
type CredentialApplication struct {
	Application Application `json:"credential_application" validate:"required"`
	// Must be present if the corresponding manifest contains a presentation_definition
	PresentationSubmission *exchange.PresentationSubmission `json:"presentation_submission,omitempty" validate:"omitempty,dive"`
}

func (ca *CredentialApplication) IsEmpty() bool {
	if ca == nil {
		return true
	}
	return reflect.DeepEqual(ca, &CredentialApplication{})
}

func (ca *CredentialApplication) IsValid() error {
	if ca.IsEmpty() {
		return errors.New("application is empty")
	}
	if err := IsValidCredentialApplication(*ca); err != nil {
		return errors.Wrap(err, "application failed json schema validation")
	}
	if ca.Application.Format != nil {
		if err := exchange.IsValidFormatDeclaration(*ca.Application.Format); err != nil {
			return errors.Wrap(err, "application's claim format failed json schema validation")
		}
	}
	return util.NewValidator().Struct(ca)
}

type Application struct {
	ID         string                `json:"id" validate:"required"`
	ManifestID string                `json:"manifest_id" validate:"required"`
	Format     *exchange.ClaimFormat `json:"format" validate:"required,dive"`
}

// CredentialFulfillment https://identity.foundation/credential-manifest/#credential-fulfillment
type CredentialFulfillment struct {
	ID            string                          `json:"id" validate:"required"`
	ManifestID    string                          `json:"manifest_id" validate:"required"`
	DescriptorMap []exchange.SubmissionDescriptor `json:"descriptor_map" validate:"required"`
}

func (cf *CredentialFulfillment) IsEmpty() bool {
	if cf == nil {
		return true
	}
	return reflect.DeepEqual(cf, &CredentialFulfillment{})
}

func (cf *CredentialFulfillment) IsValid() error {
	if cf.IsEmpty() {
		return errors.New("fulfillment is empty")
	}
	if err := IsValidCredentialFulfillment(*cf); err != nil {
		return errors.Wrap(err, "fulfillment failed json schema validation")
	}
	return util.NewValidator().Struct(cf)
}

// TODO(gabe) support multiple embed targets https://github.com/TBD54566975/ssi-sdk/issues/57
