package manifest

import (
	"fmt"
	errorresponse "github.com/TBD54566975/ssi-sdk/error"
	"reflect"
	"strings"

	"github.com/TBD54566975/ssi-sdk/credential"
	"github.com/TBD54566975/ssi-sdk/credential/exchange"
	"github.com/TBD54566975/ssi-sdk/credential/rendering"
	"github.com/TBD54566975/ssi-sdk/util"
	"github.com/goccy/go-json"
	"github.com/oliveagle/jsonpath"
	"github.com/pkg/errors"
)

const (
	CredentialManifestJSONProperty    = "credential_manifest"
	CredentialApplicationJSONProperty = "credential_application"
	CredentialResponseJSONProperty    = "credential_response"
)

// CredentialManifest https://identity.foundation/credential-manifest/#general-composition
type CredentialManifest struct {
	ID                     string                           `json:"id" validate:"required"`
	SpecVersion            string                           `json:"spec_version" validate:"required"`
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

	// validate against json schema
	if err := IsValidCredentialManifest(*cm); err != nil {
		return errors.Wrap(err, "manifest failed json schema validation")
	}

	// validate against json schema
	if err := AreValidOutputDescriptors(cm.OutputDescriptors); err != nil {
		return errors.Wrap(err, "manifest's output descriptors failed json schema validation")
	}

	// validate against struct tags
	return util.NewValidator().Struct(cm)
}

type Issuer struct {
	ID   string `json:"id" validate:"required"`
	Name string `json:"name,omitempty"`
	// an object or URI as defined by the DIF Entity Styles specification
	// https://identity.foundation/wallet-rendering/#entity-styles
	Styles *rendering.EntityStyleDescriptor `json:"styles,omitempty"`
}

// OutputDescriptor https://identity.foundation/credential-manifest/#output-descriptor
type OutputDescriptor struct {
	// Must be unique within a manifest
	ID          string `json:"id" validate:"required"`
	Schema      string `json:"schema" validate:"required"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// both below: an object or URI as defined by the DIF Entity Styles specification
	Display *rendering.DataDisplay           `json:"display,omitempty"`
	Styles  *rendering.EntityStyleDescriptor `json:"styles,omitempty"`
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

type CredentialApplicationWrapper struct {
	CredentialApplication CredentialApplication `json:"credential_application"`
	Credentials           []interface{}         `json:"verifiableCredentials,omitempty"`
}

// CredentialApplication https://identity.foundation/credential-manifest/#credential-application
type CredentialApplication struct {
	ID          string                `json:"id" validate:"required"`
	SpecVersion string                `json:"spec_version" validate:"required"`
	ManifestID  string                `json:"manifest_id" validate:"required"`
	Format      *exchange.ClaimFormat `json:"format" validate:"required,dive"`
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
	if ca.Format != nil {
		if err := exchange.IsValidDefinitionClaimFormatDesignation(*ca.Format); err != nil {
			return errors.Wrap(err, "application's claim format failed json schema validation")
		}
	}
	return util.NewValidator().Struct(ca)
}

type CredentialResponseWrapper struct {
	CredentialResponse CredentialResponse `json:"credential_response"`
	Credentials        []interface{}      `json:"verifiableCredentials,omitempty"`
}

// CredentialResponse https://identity.foundation/credential-manifest/#credential-response
type CredentialResponse struct {
	ID            string `json:"id" validate:"required"`
	SpecVersion   string `json:"spec_version" validate:"required"`
	ManifestID    string `json:"manifest_id" validate:"required"`
	ApplicationID string `json:"application_id"`
	Fulfillment   *struct {
		DescriptorMap []exchange.SubmissionDescriptor `json:"descriptor_map" validate:"required"`
	} `json:"fulfillment,omitempty" validate:"omitempty,dive"`
	Denial *struct {
		Reason           string   `json:"reason" validate:"required"`
		InputDescriptors []string `json:"input_descriptors,omitempty"`
	} `json:"denial,omitempty" validate:"omitempty,dive"`
}

func (cf *CredentialResponse) IsEmpty() bool {
	if cf == nil {
		return true
	}
	return reflect.DeepEqual(cf, &CredentialResponse{})
}

func (cf *CredentialResponse) IsValid() error {
	if cf.IsEmpty() {
		return errors.New("response is empty")
	}
	if err := IsValidCredentialResponse(*cf); err != nil {
		return errors.Wrap(err, "response failed json schema validation")
	}
	return util.NewValidator().Struct(cf)
}

// IsValidCredentialApplicationForManifest validates the rules on how a credential manifest [cm] and credential
// application [ca] relate to each other https://identity.foundation/credential-manifest/#credential-application
// applicationAndCredsJSON is the credential application and credentials as a JSON object
func IsValidCredentialApplicationForManifest(cm CredentialManifest, applicationAndCredsJSON map[string]interface{}) error {
	// parse out the application to its known object model
	applicationJSON, ok := applicationAndCredsJSON[CredentialApplicationJSONProperty]
	if !ok {
		return errorresponse.NewErrorResponse("credential_application property not found", errorresponse.ApplicationError)
	}

	applicationBytes, err := json.Marshal(applicationJSON)
	if err != nil {
		return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "failed to marshal credential application"), errorresponse.CriticalError)
	}
	var ca CredentialApplication
	if err = json.Unmarshal(applicationBytes, &ca); err != nil {
		return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "failed to unmarshal credential application"), errorresponse.CriticalError)
	}

	// Basic Validation Checks
	if err = cm.IsValid(); err != nil {
		return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "credential manifest is not valid"), errorresponse.ApplicationError)
	}

	if err = ca.IsValid(); err != nil {
		return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "credential application is not valid"), errorresponse.ApplicationError)
	}

	// The object MUST contain a manifest_id property. The value of this property MUST be the id of a valid Credential Manifest.
	if cm.ID != ca.ManifestID {
		return errorresponse.NewErrorResponse(fmt.Sprintf("the credential application's manifest id: %s must be equal to the credential manifest's id: %s", ca.ManifestID, cm.ID), errorresponse.ApplicationError)
	}

	// The ca must have a format property if the related Credential Manifest specifies a format property.
	// Its value must be a subset of the format property in the Credential Manifest that this Credential Submission
	if !cm.Format.IsEmpty() {
		cmFormats := make(map[string]bool)

		for _, format := range cm.Format.FormatValues() {
			cmFormats[format] = true
		}

		for _, format := range ca.Format.FormatValues() {
			if _, ok := cmFormats[format]; !ok {
				return errorresponse.NewErrorResponse("credential application's format must be a subset of the format property in the credential manifest", errorresponse.ApplicationError)
			}
		}
	}

	if (cm.PresentationDefinition != nil && len(cm.PresentationDefinition.InputDescriptors) > 0) &&
		(ca.PresentationSubmission == nil || len(ca.PresentationSubmission.DescriptorMap) == 0) {
		return errorresponse.NewErrorResponse(fmt.Sprintf("no descriptors provided for application: %s against manifest: %s", ca.ID, cm.ID), errorresponse.ApplicationError)
	}

	// The Credential Application object MUST contain a presentation_submission property IF the related Credential Manifest contains a presentation_definition.
	// Its value MUST be a valid Presentation Submission:
	if !cm.PresentationDefinition.IsEmpty() {
		if ca.PresentationSubmission.IsEmpty() {
			return errorresponse.NewErrorResponse("credential application's presentation submission cannot be empty because the credential manifest's presentation definition is not empty", errorresponse.ApplicationError)
		}

		if err = cm.PresentationDefinition.IsValid(); err != nil {
			return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "credential manifest's presentation definition is not valid"), errorresponse.ApplicationError)
		}

		if err = ca.PresentationSubmission.IsValid(); err != nil {
			return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "credential application's presentation submission is not valid"), errorresponse.ApplicationError)
		}

		// https://identity.foundation/presentation-exchange/#presentation-submission
		// The presentation_submission object MUST contain a definition_id property. The value of this property MUST be the id value of a valid Presentation Definition.
		if cm.PresentationDefinition.ID != ca.PresentationSubmission.DefinitionID {
			return errorresponse.NewErrorResponse(fmt.Sprintf("credential application's presentation submission's definition id: %s does not match the credential manifest's id: %s", ca.PresentationSubmission.DefinitionID, cm.PresentationDefinition.ID), errorresponse.ApplicationError)
		}

		// The descriptor_map object MUST include a format property. The value of this property MUST be a string that matches one of the Claim Format Designation. This denotes the data format of the Claim.
		claimFormats := make(map[string]bool)
		for _, format := range exchange.SupportedClaimFormats() {
			claimFormats[string(format)] = true
		}

		for _, submissionDescriptor := range ca.PresentationSubmission.DescriptorMap {
			if _, ok := claimFormats[submissionDescriptor.Format]; !ok {
				return errorresponse.NewErrorResponse("claim format is invalid or not supported", errorresponse.ApplicationError)
			}

			// The descriptor_map object MUST include a path property. The value of this property MUST be a JSONPath string expression.
			if _, err := jsonpath.Compile(submissionDescriptor.Path); err != nil {
				return errorresponse.NewErrorResponse(fmt.Sprintf("invalid json path: %s", submissionDescriptor.Path), errorresponse.ApplicationError)
			}
		}

		// index submission descriptors by id of the input descriptor
		submissionDescriptorLookup := make(map[string]exchange.SubmissionDescriptor)
		for _, d := range ca.PresentationSubmission.DescriptorMap {
			submissionDescriptorLookup[d.ID] = d
		}

		// validate each input descriptor is fulfilled
		for _, inputDescriptor := range cm.PresentationDefinition.InputDescriptors {
			submissionDescriptor, ok := submissionDescriptorLookup[inputDescriptor.ID]
			if !ok {
				return errorresponse.NewErrorResponse(fmt.Sprintf("unfulfilled input descriptor<%s>; submission not valid", inputDescriptor.ID), errorresponse.ApplicationError)
			}

			// if the format on the submitted claim does not match the input descriptor, we cannot process
			if inputDescriptor.Format != nil && !util.Contains(submissionDescriptor.Format, inputDescriptor.Format.FormatValues()) {
				return errorresponse.NewErrorResponse(fmt.Sprintf("for input descriptor<%s>, the format of submission descriptor<%s> is not one"+
					" of the supported formats: %s", inputDescriptor.ID, submissionDescriptor.Format,
					strings.Join(inputDescriptor.Format.FormatValues(), ", ")), errorresponse.ApplicationError)
			}

			// TODO(gabe) support nested paths in presentation submissions
			// https://github.com/TBD54566975/ssi-sdk/issues/73
			if submissionDescriptor.PathNested != nil {
				return fmt.Errorf("submission with nested paths not supported: %s", submissionDescriptor.ID)
				return errorresponse.NewErrorResponse(fmt.Sprintf("submission with nested paths not supported: %s", submissionDescriptor.ID), errorresponse.ApplicationError)

			}

			// resolve the claim from the JSON path expression in the submission descriptor
			submittedClaim, err := jsonpath.JsonPathLookup(applicationAndCredsJSON, submissionDescriptor.Path)
			if err != nil {
				return errorresponse.NewErrorResponseWithError(errors.Wrapf(err, "could not resolve claim from submission descriptor<%s> with path: %s", submissionDescriptor.ID, submissionDescriptor.Path), errorresponse.ApplicationError)
			}

			// convert submitted claim vc to map[string]interface{}
			var cred credential.VerifiableCredential
			credBytes, err := json.Marshal(submittedClaim)
			if err != nil {
				return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "failed to marshal submitted vc"), errorresponse.CriticalError)

			}
			if err = json.Unmarshal(credBytes, &cred); err != nil {
				return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "failed to unmarshal submitted vc"), errorresponse.CriticalError)

			}
			if err = cred.IsValid(); err != nil {
				return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "vc is not valid"), errorresponse.ApplicationError)
			}

			// verify the submitted claim complies with the input descriptor

			// if there are no constraints, we are done checking for validity
			if inputDescriptor.Constraints == nil {
				continue
			}

			// TODO(gabe) consider enforcing limited disclosure if present
			// for each field we need to verify at least one path matches
			vcMap := make(map[string]interface{})
			if err = json.Unmarshal(credBytes, &vcMap); err != nil {
				return errorresponse.NewErrorResponseWithError(errors.Wrap(err, "problem in unmarshalling credential"), errorresponse.CriticalError)

			}
			for _, field := range inputDescriptor.Constraints.Fields {
				if err = findMatchingPath(vcMap, field.Path); err != nil {
					return errorresponse.NewErrorResponseWithError(errors.Wrapf(err, "input descriptor<%s> not fulfilled for field: %s", inputDescriptor.ID, field.ID), errorresponse.ApplicationError)

				}
			}
		}
	}

	return nil
}

func findMatchingPath(claim interface{}, paths []string) error {
	for _, path := range paths {
		if _, err := jsonpath.JsonPathLookup(claim, path); err == nil {
			return nil
		}
	}
	return errors.New("matching path for claim could not be found")

}

// TODO(gabe) support multiple embed targets https://github.com/TBD54566975/ssi-sdk/issues/57
