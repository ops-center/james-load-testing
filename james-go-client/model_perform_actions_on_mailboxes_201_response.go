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

// checks if the PerformActionsOnMailboxes201Response type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PerformActionsOnMailboxes201Response{}

// PerformActionsOnMailboxes201Response struct for PerformActionsOnMailboxes201Response
type PerformActionsOnMailboxes201Response struct {
	TaskId *string `json:"taskId,omitempty"`
}

// NewPerformActionsOnMailboxes201Response instantiates a new PerformActionsOnMailboxes201Response object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPerformActionsOnMailboxes201Response() *PerformActionsOnMailboxes201Response {
	this := PerformActionsOnMailboxes201Response{}
	return &this
}

// NewPerformActionsOnMailboxes201ResponseWithDefaults instantiates a new PerformActionsOnMailboxes201Response object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPerformActionsOnMailboxes201ResponseWithDefaults() *PerformActionsOnMailboxes201Response {
	this := PerformActionsOnMailboxes201Response{}
	return &this
}

// GetTaskId returns the TaskId field value if set, zero value otherwise.
func (o *PerformActionsOnMailboxes201Response) GetTaskId() string {
	if o == nil || IsNil(o.TaskId) {
		var ret string
		return ret
	}
	return *o.TaskId
}

// GetTaskIdOk returns a tuple with the TaskId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *PerformActionsOnMailboxes201Response) GetTaskIdOk() (*string, bool) {
	if o == nil || IsNil(o.TaskId) {
		return nil, false
	}
	return o.TaskId, true
}

// HasTaskId returns a boolean if a field has been set.
func (o *PerformActionsOnMailboxes201Response) HasTaskId() bool {
	if o != nil && !IsNil(o.TaskId) {
		return true
	}

	return false
}

// SetTaskId gets a reference to the given string and assigns it to the TaskId field.
func (o *PerformActionsOnMailboxes201Response) SetTaskId(v string) {
	o.TaskId = &v
}

func (o PerformActionsOnMailboxes201Response) MarshalJSON() ([]byte, error) {
	toSerialize,err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PerformActionsOnMailboxes201Response) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.TaskId) {
		toSerialize["taskId"] = o.TaskId
	}
	return toSerialize, nil
}

type NullablePerformActionsOnMailboxes201Response struct {
	value *PerformActionsOnMailboxes201Response
	isSet bool
}

func (v NullablePerformActionsOnMailboxes201Response) Get() *PerformActionsOnMailboxes201Response {
	return v.value
}

func (v *NullablePerformActionsOnMailboxes201Response) Set(val *PerformActionsOnMailboxes201Response) {
	v.value = val
	v.isSet = true
}

func (v NullablePerformActionsOnMailboxes201Response) IsSet() bool {
	return v.isSet
}

func (v *NullablePerformActionsOnMailboxes201Response) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePerformActionsOnMailboxes201Response(val *PerformActionsOnMailboxes201Response) *NullablePerformActionsOnMailboxes201Response {
	return &NullablePerformActionsOnMailboxes201Response{value: val, isSet: true}
}

func (v NullablePerformActionsOnMailboxes201Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePerformActionsOnMailboxes201Response) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


