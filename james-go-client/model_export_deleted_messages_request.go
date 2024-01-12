/*
Apache JAMES Web Admin API

No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)

API version: 3.8.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package openapi

import (
	"encoding/json"
)

// checks if the ExportDeletedMessagesRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ExportDeletedMessagesRequest{}

// ExportDeletedMessagesRequest struct for ExportDeletedMessagesRequest
type ExportDeletedMessagesRequest struct {
	Combinator *string `json:"combinator,omitempty"`
	Criteria []Criterion `json:"criteria,omitempty"`
}

// NewExportDeletedMessagesRequest instantiates a new ExportDeletedMessagesRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewExportDeletedMessagesRequest() *ExportDeletedMessagesRequest {
	this := ExportDeletedMessagesRequest{}
	return &this
}

// NewExportDeletedMessagesRequestWithDefaults instantiates a new ExportDeletedMessagesRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewExportDeletedMessagesRequestWithDefaults() *ExportDeletedMessagesRequest {
	this := ExportDeletedMessagesRequest{}
	return &this
}

// GetCombinator returns the Combinator field value if set, zero value otherwise.
func (o *ExportDeletedMessagesRequest) GetCombinator() string {
	if o == nil || IsNil(o.Combinator) {
		var ret string
		return ret
	}
	return *o.Combinator
}

// GetCombinatorOk returns a tuple with the Combinator field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ExportDeletedMessagesRequest) GetCombinatorOk() (*string, bool) {
	if o == nil || IsNil(o.Combinator) {
		return nil, false
	}
	return o.Combinator, true
}

// HasCombinator returns a boolean if a field has been set.
func (o *ExportDeletedMessagesRequest) HasCombinator() bool {
	if o != nil && !IsNil(o.Combinator) {
		return true
	}

	return false
}

// SetCombinator gets a reference to the given string and assigns it to the Combinator field.
func (o *ExportDeletedMessagesRequest) SetCombinator(v string) {
	o.Combinator = &v
}

// GetCriteria returns the Criteria field value if set, zero value otherwise.
func (o *ExportDeletedMessagesRequest) GetCriteria() []Criterion {
	if o == nil || IsNil(o.Criteria) {
		var ret []Criterion
		return ret
	}
	return o.Criteria
}

// GetCriteriaOk returns a tuple with the Criteria field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *ExportDeletedMessagesRequest) GetCriteriaOk() ([]Criterion, bool) {
	if o == nil || IsNil(o.Criteria) {
		return nil, false
	}
	return o.Criteria, true
}

// HasCriteria returns a boolean if a field has been set.
func (o *ExportDeletedMessagesRequest) HasCriteria() bool {
	if o != nil && !IsNil(o.Criteria) {
		return true
	}

	return false
}

// SetCriteria gets a reference to the given []Criterion and assigns it to the Criteria field.
func (o *ExportDeletedMessagesRequest) SetCriteria(v []Criterion) {
	o.Criteria = v
}

func (o ExportDeletedMessagesRequest) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ExportDeletedMessagesRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Combinator) {
		toSerialize["combinator"] = o.Combinator
	}
	if !IsNil(o.Criteria) {
		toSerialize["criteria"] = o.Criteria
	}
	return toSerialize, nil
}

type NullableExportDeletedMessagesRequest struct {
	value *ExportDeletedMessagesRequest
	isSet bool
}

func (v NullableExportDeletedMessagesRequest) Get() *ExportDeletedMessagesRequest {
	return v.value
}

func (v *NullableExportDeletedMessagesRequest) Set(val *ExportDeletedMessagesRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableExportDeletedMessagesRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableExportDeletedMessagesRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableExportDeletedMessagesRequest(val *ExportDeletedMessagesRequest) *NullableExportDeletedMessagesRequest {
	return &NullableExportDeletedMessagesRequest{value: val, isSet: true}
}

func (v NullableExportDeletedMessagesRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableExportDeletedMessagesRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


